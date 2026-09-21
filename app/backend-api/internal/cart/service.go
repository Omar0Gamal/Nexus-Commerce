package cart

import (
	"context"
	"errors"
	"fmt"
	"math"

	"backend-api/internal/db"
	"backend-api/internal/orders"
	sharedanalytics "backend-api/internal/shared/analytics"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrVariantNotFound = errors.New("variant not found")
	ErrEmptyCart       = errors.New("cart is empty")
	ErrInvalidUUID     = apperr.ErrInvalidUUID
	ErrOutOfStock      = errors.New("one or more items are out of stock")
)

// Decouples cart package from concrete orders implementation.
type OrderPlacer interface {
	CreateOrder(ctx context.Context, shopID string, req orders.CreateOrderRequest) (*orders.OrderResponse, error)
}

type Service struct {
	store *Store
	q     *db.Queries
	pool  *pgxpool.Pool
	rdb   *redis.Client
}

func NewService(store *Store, q *db.Queries, pool *pgxpool.Pool, rdb *redis.Client) *Service {
	return &Service{store: store, q: q, pool: pool, rdb: rdb}
}

func (s *Service) GetCart(ctx context.Context, shopID, identity string) (*Cart, error) {
	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return nil, err
	}
	return buildCart(items), nil
}

func (s *Service) AddItem(ctx context.Context, shopID, identity string, isCustomer bool, req AddItemRequest) (*Cart, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	productUUID, err := uuid.Parse(req.ProductID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	product, err := s.q.GetProduct(ctx, db.GetProductParams{
		ID:     productUUID,
		ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product: %w", err)
	}

	price := pgutil.NumericToFloat(product.Price)
	title := product.Title
	variantName := ""
	variantID := ""

	if req.VariantID != "" {
		vUUID, err := uuid.Parse(req.VariantID)
		if err != nil {
			return nil, ErrInvalidUUID
		}
		variant, err := s.q.GetVariant(ctx, db.GetVariantParams{
			ID:     vUUID,
			ShopID: shopUUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrVariantNotFound
			}
			return nil, fmt.Errorf("get variant: %w", err)
		}
		price = pgutil.NumericToFloat(variant.Price)
		variantName = variant.Title
		variantID = req.VariantID
	}

	imageURL := ""
	images, err := s.q.ListImagesByProduct(ctx, db.ListImagesByProductParams{
		ProductID: product.ID,
		ShopID:    product.ShopID.Bytes,
	})
	if err == nil && len(images) > 0 {
		for _, img := range images {
			if img.IsPrimary.Bool {
				imageURL = img.Url
				break
			}
		}
		if imageURL == "" {
			imageURL = images[0].Url
		}
	}

	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return nil, err
	}

	found := false
	for i := range items {
		if items[i].ProductID == req.ProductID && items[i].VariantID == variantID {
			items[i].Quantity += req.Quantity
			items[i].Price = price // refresh price
			items[i].Title = title
			items[i].ImageURL = imageURL
			found = true
			break
		}
	}

	if !found {
		items = append(items, CartItem{
			ProductID:   req.ProductID,
			VariantID:   variantID,
			Title:       title,
			Price:       price,
			Quantity:    req.Quantity,
			ImageURL:    imageURL,
			Slug:        product.Slug,
			VariantName: variantName,
		})
	}

	if err := s.store.Save(ctx, shopID, identity, items, isCustomer); err != nil {
		return nil, err
	}

	if s.rdb != nil {
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "cart_add",
			ShopID: shopID,
			Payload: map[string]any{
				"product_id": req.ProductID,
				"variant_id": variantID,
				"quantity":   req.Quantity,
			},
		})
	}

	return buildCart(items), nil
}

func (s *Service) UpdateItem(ctx context.Context, shopID, identity string, isCustomer bool, productID string, variantID string, req UpdateItemRequest) (*Cart, error) {
	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return nil, err
	}

	if req.Quantity == 0 {
		filtered := make([]CartItem, 0, len(items))
		for _, item := range items {
			if !(item.ProductID == productID && item.VariantID == variantID) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	} else {
		found := false
		for i := range items {
			if items[i].ProductID == productID && items[i].VariantID == variantID {
				items[i].Quantity = req.Quantity
				found = true
				break
			}
		}
		if !found {
			return nil, ErrProductNotFound
		}
	}

	if err := s.store.Save(ctx, shopID, identity, items, isCustomer); err != nil {
		return nil, err
	}
	return buildCart(items), nil
}

func (s *Service) RemoveItem(ctx context.Context, shopID, identity string, isCustomer bool, productID string, variantID string) (*Cart, error) {
	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return nil, err
	}

	filtered := make([]CartItem, 0, len(items))
	for _, item := range items {
		if !(item.ProductID == productID && item.VariantID == variantID) {
			filtered = append(filtered, item)
		}
	}

	if err := s.store.Save(ctx, shopID, identity, filtered, isCustomer); err != nil {
		return nil, err
	}

	if s.rdb != nil {
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "cart_remove",
			ShopID: shopID,
			Payload: map[string]any{
				"product_id": productID,
				"variant_id": variantID,
			},
		})
	}

	return buildCart(filtered), nil
}

func (s *Service) ClearCart(ctx context.Context, shopID, identity string) error {
	items, _ := s.store.Get(ctx, shopID, identity)
	if err := s.store.Delete(ctx, shopID, identity); err != nil {
		return err
	}

	if s.rdb != nil && len(items) > 0 {
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "cart_abandon",
			ShopID: shopID,
			Payload: map[string]any{
				"item_count": len(items),
			},
		})
	}
	return nil
}

func (s *Service) Checkout(ctx context.Context, shopID, identity string, customerID string, req CheckoutRequest, ordersSvc OrderPlacer) (*orders.OrderResponse, error) {
	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrEmptyCart
	}

	shopUUIDForStock, _ := uuid.Parse(shopID)
	for _, item := range items {
		if item.VariantID == "" {
			continue
		}
		variantUUID, err := uuid.Parse(item.VariantID)
		if err != nil {
			continue
		}
		v, err := s.q.GetVariant(ctx, db.GetVariantParams{
			ID:     variantUUID,
			ShopID: shopUUIDForStock,
		})
		if err != nil {
			continue
		}
		if v.StockQuantity < int32(item.Quantity) {
			return nil, fmt.Errorf("%w: %s has only %d left", ErrOutOfStock, item.Title, v.StockQuantity)
		}
	}

	orderItems := make([]orders.OrderItemInput, len(items))
	for i, item := range items {
		title := item.Title
		if item.VariantName != "" {
			title = fmt.Sprintf("%s — %s", item.Title, item.VariantName)
		}
		orderItems[i] = orders.OrderItemInput{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Name:      title,
			Price:     item.Price,
			Quantity:  int32(item.Quantity),
		}
	}

	orderReq := orders.CreateOrderRequest{
		CustomerID:    customerID,
		CustomerEmail: req.Email,
		ShippingAddress: orders.ShippingAddress{
			FirstName: req.ShippingAddress.FirstName,
			LastName:  req.ShippingAddress.LastName,
			Address1:  req.ShippingAddress.Address1,
			Address2:  req.ShippingAddress.Address2,
			City:      req.ShippingAddress.City,
			State:     req.ShippingAddress.State,
			Country:   req.ShippingAddress.Country,
			ZipCode:   req.ShippingAddress.ZipCode,
			Phone:     req.ShippingAddress.Phone,
		},
		Items: orderItems,
	}

	order, err := ordersSvc.CreateOrder(ctx, shopID, orderReq)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	_ = s.store.Delete(ctx, shopID, identity)

	return order, nil
}

func buildCart(items []CartItem) *Cart {
	totalItems := 0
	totalPrice := 0.0
	for _, item := range items {
		totalItems += item.Quantity
		totalPrice += item.Price * float64(item.Quantity)
	}
	totalPrice = math.Round(totalPrice*100) / 100

	return &Cart{
		Items:      items,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}
}

func (s *Service) GetCartForCheckout(ctx context.Context, shopID, identity string) ([]orders.OrderItemInput, error) {
	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrEmptyCart
	}
	orderItems := make([]orders.OrderItemInput, len(items))
	for i, item := range items {
		title := item.Title
		if item.VariantName != "" {
			title = fmt.Sprintf("%s — %s", item.Title, item.VariantName)
		}
		orderItems[i] = orders.OrderItemInput{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Name:      title,
			Price:     item.Price,
			Quantity:  int32(item.Quantity),
		}
	}
	return orderItems, nil
}

