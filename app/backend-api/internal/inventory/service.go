package inventory

import (
	"context"
	"fmt"

	"backend-api/internal/db"
	"backend-api/internal/email"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// Service handles advanced inventory operations.
type Service struct {
	queries *db.Queries
	mailer  *email.Mailer
	logger  *zap.Logger
}

func NewService(queries *db.Queries, mailer *email.Mailer, logger *zap.Logger) *Service {
	return &Service{queries: queries, mailer: mailer, logger: logger}
}

// AdjustStock adjusts the quantity in a specific warehouse and records a movement.
func (s *Service) AdjustStock(ctx context.Context, shopID, variantID, warehouseID uuid.UUID, delta int, movementType string, referenceID *uuid.UUID, referenceType, note, createdBy string) error {
	// Resolve low-stock threshold from the active reorder rule for this variant.
	threshold := int32(5)
	if rules, err := s.queries.GetActiveReorderRules(ctx, shopID); err == nil {
		for _, rule := range rules {
			if rule.VariantID == variantID {
				threshold = rule.ReorderPoint
				break
			}
		}
	}

	// Adjust the inventory level.
	_, err := s.queries.AdjustInventoryLevel(ctx, db.AdjustInventoryLevelParams{
		ShopID:            shopID,
		VariantID:         variantID,
		WarehouseID:       warehouseID,
		QuantityOnHand:    int32(delta),
		LowStockThreshold: threshold,
	})
	if err != nil {
		return fmt.Errorf("inventory.AdjustStock: %w", err)
	}

	// Record the stock movement.
	refID := pgtype.UUID{}
	if referenceID != nil {
		refID = pgtype.UUID{Bytes: *referenceID, Valid: true}
	}
	refType := pgtype.Text{}
	if referenceType != "" {
		refType = pgtype.Text{String: referenceType, Valid: true}
	}
	noteText := pgtype.Text{}
	if note != "" {
		noteText = pgtype.Text{String: note, Valid: true}
	}
	createdByText := pgtype.Text{}
	if createdBy != "" {
		createdByText = pgtype.Text{String: createdBy, Valid: true}
	}
	warehouseIDNullable := pgtype.UUID{Bytes: warehouseID, Valid: true}

	_, err = s.queries.RecordStockMovement(ctx, db.RecordStockMovementParams{
		ShopID:        shopID,
		VariantID:     variantID,
		WarehouseID:   warehouseIDNullable,
		MovementType:  movementType,
		QuantityDelta: int32(delta),
		ReferenceID:   refID,
		ReferenceType: refType,
		Note:          noteText,
		CreatedBy:     createdByText,
	})
	return err
}

// GetLevelsByVariant returns all warehouse inventory levels for a variant.
func (s *Service) GetLevelsByVariant(ctx context.Context, shopID, variantID uuid.UUID) ([]db.InventoryLevel, error) {
	levels, err := s.queries.GetInventoryLevelsByVariant(ctx, db.GetInventoryLevelsByVariantParams{
		ShopID:    shopID,
		VariantID: variantID,
	})
	if err != nil {
		return nil, err
	}
	if levels == nil {
		return []db.InventoryLevel{}, nil
	}
	return levels, nil
}

// ComputeBundleAvailability returns how many of the bundle can be sold
// based on the minimum available component quantity divided by bundle quantity.
func (s *Service) ComputeBundleAvailability(ctx context.Context, shopID, bundleProductID uuid.UUID) (int, error) {
	components, err := s.queries.GetProductBundleComponents(ctx, db.GetProductBundleComponentsParams{
		ShopID:          shopID,
		BundleProductID: bundleProductID,
	})
	if err != nil {
		return 0, err
	}
	if len(components) == 0 {
		return 0, nil
	}

	minAvailable := -1
	for _, comp := range components {
		levels, err := s.queries.GetInventoryLevelsByVariant(ctx, db.GetInventoryLevelsByVariantParams{
			ShopID:    shopID,
			VariantID: comp.ComponentVariantID,
		})
		if err != nil {
			return 0, err
		}
		totalQty := 0
		for _, lv := range levels {
			totalQty += int(lv.QuantityOnHand)
		}
		available := totalQty / int(comp.Quantity)
		if minAvailable < 0 || available < minAvailable {
			minAvailable = available
		}
	}
	if minAvailable < 0 {
		return 0, nil
	}
	return minAvailable, nil
}

// CheckReorderRules evaluates all active reorder rules for a shop and sends
// reorder notification emails when stock falls at or below the reorder point.
func (s *Service) CheckReorderRules(ctx context.Context, shopID uuid.UUID) error {
	rules, err := s.queries.GetActiveReorderRules(ctx, shopID)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		levels, err := s.queries.GetInventoryLevelsByVariant(ctx, db.GetInventoryLevelsByVariantParams{
			ShopID:    shopID,
			VariantID: rule.VariantID,
		})
		if err != nil {
			s.logger.Error("get inventory levels for reorder rule",
				zap.String("rule_id", rule.ID.String()),
				zap.Error(err))
			continue
		}
		totalQty := 0
		for _, lv := range levels {
			totalQty += int(lv.QuantityOnHand)
		}
		if totalQty > int(rule.ReorderPoint) {
			continue
		}

		// Trigger: send email and update rule.
		if rule.SupplierEmail.Valid && rule.SupplierEmail.String != "" && s.mailer != nil {
			if err := s.mailer.SendLowStockAlert(rule.SupplierEmail.String, email.LowStockData{
				ProductID:    rule.VariantID.String(),
				CurrentStock: int32(totalQty),
				Threshold:    rule.ReorderPoint,
			}); err != nil {
				s.logger.Error("send reorder email", zap.Error(err))
			}
		}

		if err := s.queries.MarkReorderRuleTriggered(ctx, rule.ID); err != nil {
			s.logger.Error("mark reorder rule triggered",
				zap.String("rule_id", rule.ID.String()),
				zap.Error(err))
		}
	}
	return nil
}


// GetWarehouses returns all warehouses for a shop.
func (s *Service) GetWarehouses(ctx context.Context, shopID uuid.UUID) ([]db.Warehouse, error) {
	rows, err := s.queries.GetWarehouses(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.Warehouse{}, nil
	}
	return rows, nil
}

// CreateWarehouse creates a new warehouse.
func (s *Service) CreateWarehouse(ctx context.Context, shopID uuid.UUID, name, address string, isDefault bool) (db.Warehouse, error) {
	addr := pgtype.Text{}
	if address != "" {
		addr = pgtype.Text{String: address, Valid: true}
	}
	return s.queries.CreateWarehouse(ctx, db.CreateWarehouseParams{
		ShopID:    shopID,
		Name:      name,
		Address:   addr,
		IsDefault: isDefault,
	})
}

// GetLowStockVariants returns variants below their low-stock threshold.
func (s *Service) GetLowStockVariants(ctx context.Context, shopID uuid.UUID) ([]db.GetLowStockVariantsRow, error) {
	rows, err := s.queries.GetLowStockVariants(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.GetLowStockVariantsRow{}, nil
	}
	return rows, nil
}

// GetStockMovements returns paginated stock movements for a variant.
func (s *Service) GetStockMovements(ctx context.Context, shopID, variantID uuid.UUID, limit, offset int32) ([]db.StockMovement, error) {
	rows, err := s.queries.GetStockMovements(ctx, db.GetStockMovementsParams{
		ShopID:    shopID,
		VariantID: variantID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.StockMovement{}, nil
	}
	return rows, nil
}

// GetProductBundleComponents returns all components for a bundle product.
func (s *Service) GetProductBundleComponents(ctx context.Context, shopID, productID uuid.UUID) ([]db.ProductBundle, error) {
	rows, err := s.queries.GetProductBundleComponents(ctx, db.GetProductBundleComponentsParams{
		ShopID:          shopID,
		BundleProductID: productID,
	})
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ProductBundle{}, nil
	}
	return rows, nil
}

// UpsertBundleComponent creates or updates a bundle component.
func (s *Service) UpsertBundleComponent(ctx context.Context, shopID, bundleProductID, componentVariantID uuid.UUID, qty int16) (db.ProductBundle, error) {
	return s.queries.UpsertBundleComponent(ctx, db.UpsertBundleComponentParams{
		ShopID:             shopID,
		BundleProductID:    bundleProductID,
		ComponentVariantID: componentVariantID,
		Quantity:           qty,
	})
}

// GetReorderRules returns all active reorder rules for a shop.
func (s *Service) GetReorderRules(ctx context.Context, shopID uuid.UUID) ([]db.ReorderRule, error) {
	rows, err := s.queries.GetActiveReorderRules(ctx, shopID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []db.ReorderRule{}, nil
	}
	return rows, nil
}

// UpsertReorderRule creates or updates a reorder rule.
func (s *Service) UpsertReorderRule(ctx context.Context, shopID, variantID uuid.UUID, reorderPoint, reorderQty int32, supplierEmail string) (db.ReorderRule, error) {
	se := pgtype.Text{}
	if supplierEmail != "" {
		se = pgtype.Text{String: supplierEmail, Valid: true}
	}
	return s.queries.UpsertReorderRule(ctx, db.UpsertReorderRuleParams{
		ShopID:        shopID,
		VariantID:     variantID,
		ReorderPoint:  reorderPoint,
		ReorderQty:    reorderQty,
		SupplierEmail: se,
	})
}
