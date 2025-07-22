package wishlists

import (
	"context"
	"errors"
	"fmt"

	"backend-api/internal/db"
	"backend-api/internal/shared/apperr"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound    = apperr.ErrNotFound
	ErrInvalidUUID = apperr.ErrInvalidUUID
)

// Service handles wishlist business logic.
type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

// Add adds a product to the customer's wishlist.
// Silently ignores duplicate additions (ON CONFLICT DO NOTHING).
func (s *Service) Add(ctx context.Context, shopID, customerID, productID string, variantID *string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return ErrInvalidUUID
	}

	var varUUID pgtype.UUID
	if variantID != nil {
		vid, err := uuid.Parse(*variantID)
		if err != nil {
			return ErrInvalidUUID
		}
		varUUID = pgtype.UUID{Bytes: vid, Valid: true}
	}

	_, err = s.q.AddToWishlist(ctx, db.AddToWishlistParams{
		ShopID:     shopUUID,
		CustomerID: customerUUID,
		ProductID:  productUUID,
		VariantID:  varUUID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("add to wishlist: %w", err)
	}
	return nil
}

// Remove removes a product from the customer's wishlist.
func (s *Service) Remove(ctx context.Context, shopID, customerID, productID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return ErrInvalidUUID
	}

	return s.q.RemoveFromWishlist(ctx, db.RemoveFromWishlistParams{
		ShopID:     shopUUID,
		CustomerID: customerUUID,
		ProductID:  productUUID,
	})
}

// Get returns all wishlist items for a customer.
func (s *Service) Get(ctx context.Context, shopID, customerID string) ([]WishlistItem, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	rows, err := s.q.GetWishlist(ctx, db.GetWishlistParams{
		ShopID:     shopUUID,
		CustomerID: customerUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("get wishlist: %w", err)
	}

	items := make([]WishlistItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, wishlistRowToItem(r))
	}
	return items, nil
}

// Check returns whether a product is in the customer's wishlist.
func (s *Service) Check(ctx context.Context, shopID, customerID, productID string) (bool, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return false, ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return false, ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return false, ErrInvalidUUID
	}

	return s.q.IsWishlisted(ctx, db.IsWishlistedParams{
		ShopID:     shopUUID,
		CustomerID: customerUUID,
		ProductID:  productUUID,
	})
}

// GetWishlistedCustomers returns customers who wishlisted a product (for back-in-stock notifications).
func (s *Service) GetWishlistedCustomers(ctx context.Context, shopID, productID string) ([]db.GetWishlistedCustomersForProductRow, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	return s.q.GetWishlistedCustomersForProduct(ctx, db.GetWishlistedCustomersForProductParams{
		ShopID:    shopUUID,
		ProductID: productUUID,
	})
}

// GetMostWishlisted returns the top N most-wishlisted product IDs for a shop.
func (s *Service) GetMostWishlisted(ctx context.Context, shopID string, limit int32) ([]db.GetMostWishlistedProductsRow, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	return s.q.GetMostWishlistedProducts(ctx, db.GetMostWishlistedProductsParams{
		ShopID: shopUUID,
		Limit:  limit,
	})
}


func wishlistRowToItem(r db.GetWishlistRow) WishlistItem {
	item := WishlistItem{
		ID:           r.ID.String(),
		ProductID:    r.ProductID.String(),
		ProductTitle: r.ProductTitle,
		ProductSlug:  r.ProductSlug,
		CreatedAt:    r.CreatedAt.Time,
	}
	if r.VariantID.Valid {
		s := uuid.UUID(r.VariantID.Bytes).String()
		item.VariantID = &s
	}
	if r.ProductPrice.Valid {
		f, _ := r.ProductPrice.Float64Value()
		item.ProductPrice = f.Float64
	}
	item.ProductStock = r.ProductStock
	return item
}
