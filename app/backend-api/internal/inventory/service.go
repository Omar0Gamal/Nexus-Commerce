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

type Service struct {
	queries *db.Queries
	mailer  *email.Mailer
	logger  *zap.Logger
}

func NewService(queries *db.Queries, mailer *email.Mailer, logger *zap.Logger) *Service {
	return &Service{queries: queries, mailer: mailer, logger: logger}
}

func (s *Service) AdjustStock(ctx context.Context, shopID, variantID, warehouseID uuid.UUID, delta int, movementType string, referenceID *uuid.UUID, referenceType, note, createdBy string) error {
	threshold := int32(5)
	if rules, err := s.queries.GetActiveReorderRules(ctx, shopID); err == nil {
		for _, rule := range rules {
			if rule.VariantID == variantID {
				threshold = rule.ReorderPoint
				break
			}
		}
	}

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

func (s *Service) UpsertBundleComponent(ctx context.Context, shopID, bundleProductID, componentVariantID uuid.UUID, qty int16) (db.ProductBundle, error) {
	return s.queries.UpsertBundleComponent(ctx, db.UpsertBundleComponentParams{
		ShopID:             shopID,
		BundleProductID:    bundleProductID,
		ComponentVariantID: componentVariantID,
		Quantity:           qty,
	})
}

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
