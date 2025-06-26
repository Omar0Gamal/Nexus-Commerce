package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"backend-api/internal/coupons"
	"backend-api/internal/db"
	"backend-api/internal/email"
	sharedanalytics "backend-api/internal/shared/analytics"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"
	"backend-api/internal/shipping"
	"backend-api/internal/worker"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Sentinel errors
var (
	ErrNotFound          = apperr.ErrNotFound
	ErrInvalidUUID       = apperr.ErrInvalidUUID
	ErrBadStatus         = errors.New("invalid order status")
	ErrInsufficientStock = errors.New("insufficient stock for one or more items")
)

// validStatuses for the order_status enum.
var validStatuses = map[string]bool{
	"pending":    true,
	"paid":       true,
	"processing": true,
	"shipped":    true,
	"completed":  true,
	"cancelled":  true,
	"refunded":   true,
}

// WebhookDispatcher is satisfied by *webhooks.Service and used to fire
// webhook events without creating an import cycle.
type WebhookDispatcher interface {
	Dispatch(shopID, topic string, data map[string]any)
}

// ServiceDeps holds dependencies for the orders service.
type ServiceDeps struct {
	Coupons  *coupons.Service
	Shipping *shipping.Service
	Mailer   *email.Mailer
	Logger   *zap.Logger
	Webhooks WebhookDispatcher
	Jobs     *worker.Queue
	Redis    *redis.Client
}

// Service handles order business logic.
type Service struct {
	q        *db.Queries
	pool     *pgxpool.Pool
	coupons  *coupons.Service
	shipping *shipping.Service
	mailer   *email.Mailer
	logger   *zap.Logger
	webhooks WebhookDispatcher
	jobs     *worker.Queue
	rdb      *redis.Client
}

func NewService(q *db.Queries, pool *pgxpool.Pool, deps ServiceDeps) *Service {
	logger := deps.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		q:        q,
		pool:     pool,
		coupons:  deps.Coupons,
		shipping: deps.Shipping,
		mailer:   deps.Mailer,
		logger:   logger,
		webhooks: deps.Webhooks,
		jobs:     deps.Jobs,
		rdb:      deps.Redis,
	}
}

// sendLowStockAlert fires a non-blocking low-stock email to the shop owner
// and a product.low_stock webhook dispatch.
func (s *Service) sendLowStockAlert(shopUUID uuid.UUID, row db.DecrementProductStockRow) {
	shopID := uuid.UUID(shopUUID).String()
	productID := row.ID
	stockQty := row.StockQuantity
	threshold := row.LowStockThreshold

	// Webhook dispatch (does not need mailer)
	if s.webhooks != nil {
		s.webhooks.Dispatch(shopID, "product.low_stock", map[string]any{
			"product_id": productID.String(),
			"stock":      stockQty,
			"threshold":  threshold,
		})
	}

	if s.mailer == nil && s.jobs == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		shop, err := s.q.GetShop(ctx, shopUUID)
		if err != nil {
			return
		}
		owner, err := s.q.GetUser(ctx, shop.OwnerUserID)
		if err != nil {
			return
		}
		product, err := s.q.GetProduct(ctx, db.GetProductParams{
			ID:     productID,
			ShopID: pgtype.UUID{Bytes: shopUUID, Valid: true},
		})
		if err != nil {
			return
		}
		data := email.LowStockData{
			ShopName:     shop.Name,
			ProductTitle: product.Title,
			ProductID:    product.ID.String(),
			SKU:          product.Sku.String,
			CurrentStock: stockQty,
			Threshold:    threshold,
		}
		if s.jobs != nil {
			_ = s.jobs.EnqueueTyped(ctx, worker.JobSendLowStockAlert,
				worker.LowStockAlertPayload{To: owner.Email, Data: data})
		} else if s.mailer != nil {
			if err := s.mailer.SendLowStockAlert(owner.Email, data); err != nil {
				s.logger.Error("send low-stock alert email", zap.Error(err))
			}
		}
	}()
}

func (s *Service) CreateOrder(ctx context.Context, shopID string, req CreateOrderRequest) (*OrderResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	// Marshal shipping address
	addrJSON, err := json.Marshal(req.ShippingAddress)
	if err != nil {
		return nil, fmt.Errorf("marshal shipping address: %w", err)
	}

	// Calculate total price from items using integer-cent arithmetic
	var totalCents int64
	for _, item := range req.Items {
		priceCents := int64(math.Round(item.Price * 100))
		totalCents += priceCents * int64(item.Quantity)
	}

	var couponResult *coupons.ApplyResult
	if req.CouponCode != "" && s.coupons != nil {
		subtotal := float64(totalCents) / 100
		cr, err := s.coupons.ApplyCoupon(ctx, shopID, req.CouponCode, subtotal)
		if err != nil {
			return nil, fmt.Errorf("coupon: %w", err)
		}
		couponResult = cr
		totalCents -= int64(math.Round(cr.DiscountAmount * 100))
		if totalCents < 0 {
			totalCents = 0
		}
	}

	var shippingFeeCents int64
	if s.shipping != nil && req.ShippingAddress.Country != "" {
		subtotalAfterDiscount := float64(totalCents) / 100
		fee, err := s.shipping.ResolveShippingFee(ctx, shopID, req.ShippingAddress.Country, subtotalAfterDiscount, req.WeightKg)
		if err == nil {
			shippingFeeCents = int64(math.Round(fee * 100))
			totalCents += shippingFeeCents
		}
		// If no rate found (ErrNoRate), continue without shipping fee rather than failing
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			s.logger.Warn("order transaction rollback failed", zap.Error(rbErr))
		}
	}()

	qtx := s.q.WithTx(tx)

	// Get next order number (inside tx for consistency)
	nextNum, err := qtx.GetNextOrderNumber(ctx, shopUUID)
	if err != nil {
		return nil, fmt.Errorf("get next order number: %w", err)
	}

	// Build order params
	params := db.CreateOrderParams{
		ShopID:          shopUUID,
		OrderNumber:     nextNum,
		TotalPrice:      pgtype.Numeric{Int: big.NewInt(totalCents), Exp: -2, Valid: true},
		DiscountAmount:  pgtype.Numeric{Int: big.NewInt(0), Exp: -2, Valid: true},
		ShippingFee:     pgtype.Numeric{Int: big.NewInt(shippingFeeCents), Exp: -2, Valid: true},
		Status:          db.NullOrderStatus{OrderStatus: db.OrderStatusPending, Valid: true},
		ShippingAddress: addrJSON,
	}

	if couponResult != nil {
		params.CouponID = pgutil.ToUUID(couponResult.CouponID)
		discountCents := int64(math.Round(couponResult.DiscountAmount * 100))
		params.DiscountAmount = pgtype.Numeric{Int: big.NewInt(discountCents), Exp: -2, Valid: true}
	}

	if req.CustomerID != "" {
		custUUID, err := uuid.Parse(req.CustomerID)
		if err != nil {
			return nil, fmt.Errorf("customer_id: %w", ErrInvalidUUID)
		}
		params.CustomerID = pgutil.ToUUID(custUUID)
	}

	order, err := qtx.CreateOrder(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Create order items + atomically decrement stock
	var items []OrderItemResponse
	for _, item := range req.Items {
		itemParams := db.CreateOrderItemParams{
			OrderID:  pgutil.ToUUID(order.ID),
			Name:     item.Name,
			Price:    pgutil.FloatToNumericCents(item.Price),
			Quantity: item.Quantity,
		}

		var prodUUID uuid.UUID
		hasProd := false
		if item.ProductID != "" {
			prodUUID, err = uuid.Parse(item.ProductID)
			if err != nil {
				return nil, fmt.Errorf("product_id: %w", ErrInvalidUUID)
			}
			itemParams.ProductID = pgutil.ToUUID(prodUUID)
			hasProd = true
		}
		var varUUID uuid.UUID
		hasVar := false
		if item.VariantID != "" {
			varUUID, err = uuid.Parse(item.VariantID)
			if err != nil {
				return nil, fmt.Errorf("variant_id: %w", ErrInvalidUUID)
			}
			itemParams.VariantID = pgutil.ToUUID(varUUID)
			hasVar = true
		}

		oi, err := qtx.CreateOrderItem(ctx, itemParams)
		if err != nil {
			return nil, fmt.Errorf("create order item: %w", err)
		}
		items = append(items, orderItemToResponse(oi))

		if hasVar {
			_, err = qtx.DecrementVariantStock(ctx, db.DecrementVariantStockParams{
				ID:            varUUID,
				ShopID:        shopUUID,
				StockQuantity: item.Quantity,
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, fmt.Errorf("%w: variant %s", ErrInsufficientStock, item.VariantID)
				}
				return nil, fmt.Errorf("decrement variant stock: %w", err)
			}
		} else if hasProd {
			row, err := qtx.DecrementProductStock(ctx, db.DecrementProductStockParams{
				ID:            prodUUID,
				ShopID:        pgutil.ToUUID(shopUUID),
				StockQuantity: item.Quantity,
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil, fmt.Errorf("%w: product %s", ErrInsufficientStock, item.ProductID)
				}
				return nil, fmt.Errorf("decrement product stock: %w", err)
			}
			// Log low-stock warning when remaining stock hits the threshold.
			if row.StockQuantity <= row.LowStockThreshold {
				s.logger.Warn("low stock",
					zap.String("product_id", prodUUID.String()),
					zap.Int32("remaining", row.StockQuantity),
					zap.Int32("threshold", row.LowStockThreshold),
				)
				s.sendLowStockAlert(shopUUID, row)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Increment coupon usage post-commit (best-effort, inside original ctx)
	if couponResult != nil && s.coupons != nil {
		s.coupons.IncrementUsage(ctx, shopID, couponResult.CouponID)
	}

	resp := orderToResponse(order)
	resp.Items = items

	// Send order confirmation email (best-effort; never fails the request)
	if req.CustomerEmail != "" {
		emailItems := make([]email.OrderItemData, len(items))
		for i, it := range items {
			emailItems[i] = email.OrderItemData{
				Name:     it.Name,
				Quantity: it.Quantity,
				Price:    it.Price,
			}
		}
		emailData := email.OrderData{
			OrderNumber:  resp.OrderNumber,
			OrderID:      resp.ID,
			TotalPrice:   resp.TotalPrice,
			Status:       resp.Status,
			CustomerName: req.ShippingAddress.FirstName,
			Items:        emailItems,
		}
		if s.jobs != nil {
			// Enqueue to background worker (non-blocking)
			_ = s.jobs.EnqueueTyped(ctx, worker.JobSendOrderConfirmation, worker.OrderConfirmationPayload{
				To:   req.CustomerEmail,
				Data: emailData,
			})
		} else if s.mailer != nil {
			// Fallback: direct goroutine
			go func() {
				if err := s.mailer.SendOrderConfirmation(req.CustomerEmail, emailData); err != nil {
					s.logger.Error("send order confirmation email", zap.Error(err))
				}
			}()
		}
	}

	// Dispatch order.created webhook (best-effort)
	if s.webhooks != nil {
		s.webhooks.Dispatch(shopID, "order.created", map[string]any{
			"order_id":     resp.ID,
			"order_number": resp.OrderNumber,
			"total_price":  resp.TotalPrice,
			"status":       resp.Status,
		})
	}

	// Publish analytics checkout_start event (fire-and-forget)
	if s.rdb != nil {
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "checkout_start",
			ShopID: shopID,
			Payload: map[string]any{
				"order_id":    resp.ID,
				"item_count":  len(req.Items),
				"total_cents": totalCents,
			},
		})
	}

	return resp, nil
}

func (s *Service) GetOrder(ctx context.Context, shopID, orderID string) (*OrderResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	ordUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	order, err := s.q.GetOrder(ctx, db.GetOrderParams{
		ID:     ordUUID,
		ShopID: shopUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	// Load items
	items, err := s.q.ListOrderItems(ctx, pgutil.ToUUID(order.ID))
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}

	resp := orderToResponse(order)
	resp.Items = make([]OrderItemResponse, len(items))
	for i, oi := range items {
		resp.Items[i] = orderItemToResponse(oi)
	}
	return resp, nil
}

func (s *Service) ListOrders(ctx context.Context, shopID string, params ListOrdersParams) ([]OrderResponse, int64, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, 0, ErrInvalidUUID
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	offset := (params.Page - 1) * params.PerPage

	var orders []db.Order
	var total int64

	if params.Status != "" {
		if !validStatuses[params.Status] {
			return nil, 0, ErrBadStatus
		}
		orders, err = s.q.ListOrdersByStatus(ctx, db.ListOrdersByStatusParams{
			ShopID: shopUUID,
			Status: db.NullOrderStatus{OrderStatus: db.OrderStatus(params.Status), Valid: true},
			Limit:  int32(params.PerPage),
			Offset: int32(offset),
		})
		if err != nil {
			return nil, 0, fmt.Errorf("list orders by status: %w", err)
		}
		total, err = s.q.CountOrdersByStatus(ctx, db.CountOrdersByStatusParams{
			ShopID: shopUUID,
			Status: db.NullOrderStatus{OrderStatus: db.OrderStatus(params.Status), Valid: true},
		})
	} else if params.CustomerID != "" {
		custUUID, err := uuid.Parse(params.CustomerID)
		if err != nil {
			return nil, 0, ErrInvalidUUID
		}
		orders, err = s.q.ListOrdersByCustomer(ctx, db.ListOrdersByCustomerParams{
			ShopID:     shopUUID,
			CustomerID: pgutil.ToUUID(custUUID),
		})
		if err != nil {
			return nil, 0, fmt.Errorf("list orders by customer: %w", err)
		}
		total = int64(len(orders))
	} else {
		orders, err = s.q.ListOrders(ctx, db.ListOrdersParams{
			ShopID: shopUUID,
			Limit:  int32(params.PerPage),
			Offset: int32(offset),
		})
		if err != nil {
			return nil, 0, fmt.Errorf("list orders: %w", err)
		}
		total, err = s.q.CountOrders(ctx, shopUUID)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	responses := make([]OrderResponse, len(orders))
	for i, o := range orders {
		responses[i] = *orderToResponse(o)
	}
	return responses, total, nil
}

func (s *Service) UpdateOrderStatus(ctx context.Context, shopID, orderID string, status string) (*OrderResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	ordUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	if !validStatuses[status] {
		return nil, ErrBadStatus
	}

	order, err := s.q.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     ordUUID,
		Status: db.NullOrderStatus{OrderStatus: db.OrderStatus(status), Valid: true},
		ShopID: shopUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update order status: %w", err)
	}

	resp := orderToResponse(order)

	// Send status-update email to the customer if one is registered (best-effort)
	if order.CustomerID.Valid {
		custUUID := uuid.UUID(order.CustomerID.Bytes)
		orderShopID := pgtype.UUID{Bytes: order.ShopID, Valid: true}
		respCopy := *resp

		if s.jobs != nil {
			// Fetch customer inline to get email, then enqueue (goroutine to keep path non-blocking)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				customer, err := s.q.GetCustomer(ctx, db.GetCustomerParams{
					ID:     custUUID,
					ShopID: orderShopID,
				})
				if err != nil {
					s.logger.Error("lookup customer for status email", zap.Error(err))
					return
				}
				name := customer.Email
				if customer.FirstName.Valid && customer.FirstName.String != "" {
					name = customer.FirstName.String
				}
				data := email.OrderData{
					OrderNumber:  respCopy.OrderNumber,
					OrderID:      respCopy.ID,
					TotalPrice:   respCopy.TotalPrice,
					Status:       status,
					CustomerName: name,
				}
				_ = s.jobs.EnqueueTyped(ctx, worker.JobSendOrderStatusUpdate,
					worker.OrderStatusUpdatePayload{To: customer.Email, Data: data})
			}()
		} else if s.mailer != nil {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				customer, err := s.q.GetCustomer(ctx, db.GetCustomerParams{
					ID:     custUUID,
					ShopID: orderShopID,
				})
				if err != nil {
					s.logger.Error("lookup customer for status email", zap.Error(err))
					return
				}
				name := customer.Email
				if customer.FirstName.Valid && customer.FirstName.String != "" {
					name = customer.FirstName.String
				}
				data := email.OrderData{
					OrderNumber:  respCopy.OrderNumber,
					OrderID:      respCopy.ID,
					TotalPrice:   respCopy.TotalPrice,
					Status:       status,
					CustomerName: name,
				}
				if err := s.mailer.SendOrderStatusUpdate(customer.Email, data); err != nil {
					s.logger.Error("send order status update email", zap.Error(err))
				}
			}()
		}
	}

	// Dispatch order.status_changed webhook (best-effort)
	if s.webhooks != nil {
		s.webhooks.Dispatch(shopID, "order.status_changed", map[string]any{
			"order_id":     resp.ID,
			"order_number": resp.OrderNumber,
			"status":       status,
		})
	}

	// Publish analytics event for completed orders (fire-and-forget)
	if s.rdb != nil && status == "completed" {
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "order_completed",
			ShopID: shopID,
			Payload: map[string]any{
				"order_id": resp.ID,
			},
		})
	}

	return resp, nil
}

// ListMyOrders returns all orders placed by the authenticated customer.
func (s *Service) ListMyOrders(ctx context.Context, shopID, customerID string) ([]OrderResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	orders, err := s.q.ListOrdersByCustomer(ctx, db.ListOrdersByCustomerParams{
		ShopID:     shopUUID,
		CustomerID: pgutil.ToUUID(custUUID),
	})
	if err != nil {
		return nil, fmt.Errorf("list orders by customer: %w", err)
	}

	result := make([]OrderResponse, len(orders))
	for i, o := range orders {
		resp := orderToResponse(o)
		items, err := s.q.ListOrderItems(ctx, pgutil.ToUUID(o.ID))
		if err == nil {
			resp.Items = make([]OrderItemResponse, len(items))
			for j, oi := range items {
				resp.Items[j] = orderItemToResponse(oi)
			}
		}
		result[i] = *resp
	}
	return result, nil
}

// GetMyOrder returns a single order, verifying it belongs to the customer.
func (s *Service) GetMyOrder(ctx context.Context, shopID, customerID, orderID string) (*OrderResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	ordUUID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	order, err := s.q.GetOrder(ctx, db.GetOrderParams{
		ID:     ordUUID,
		ShopID: shopUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	// Verify the order belongs to this customer (return 404 to avoid leaking existence)
	if !order.CustomerID.Valid || uuid.UUID(order.CustomerID.Bytes) != custUUID {
		return nil, ErrNotFound
	}

	items, err := s.q.ListOrderItems(ctx, pgutil.ToUUID(order.ID))
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	resp := orderToResponse(order)
	resp.Items = make([]OrderItemResponse, len(items))
	for i, oi := range items {
		resp.Items[i] = orderItemToResponse(oi)
	}
	return resp, nil
}

func (s *Service) DeleteOrder(ctx context.Context, shopID, orderID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	ordUUID, err := uuid.Parse(orderID)
	if err != nil {
		return ErrInvalidUUID
	}

	return s.q.DeleteOrder(ctx, db.DeleteOrderParams{
		ID:     ordUUID,
		ShopID: shopUUID,
	})
}

func orderToResponse(o db.Order) *OrderResponse {
	resp := &OrderResponse{
		ID:              o.ID.String(),
		ShopID:          o.ShopID.String(),
		OrderNumber:     o.OrderNumber,
		TotalPrice:      pgutil.NumericToString(o.TotalPrice),
		DiscountAmount:  pgutil.NumericToString(o.DiscountAmount),
		ShippingFee:     pgutil.NumericToString(o.ShippingFee),
		Status:          string(o.Status.OrderStatus),
		ShippingAddress: o.ShippingAddress,
	}

	if o.CustomerID.Valid {
		s := uuid.UUID(o.CustomerID.Bytes).String()
		resp.CustomerID = &s
	}
	if o.CouponID.Valid {
		s := uuid.UUID(o.CouponID.Bytes).String()
		resp.CouponID = &s
	}
	if o.CreatedAt.Valid {
		resp.CreatedAt = o.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	return resp
}

func orderItemToResponse(oi db.OrderItem) OrderItemResponse {
	resp := OrderItemResponse{
		ID:       oi.ID.String(),
		OrderID:  pgutil.UUIDToString(oi.OrderID),
		Name:     oi.Name,
		Price:    pgutil.NumericToString(oi.Price),
		Quantity: oi.Quantity,
	}

	if oi.ProductID.Valid {
		s := uuid.UUID(oi.ProductID.Bytes).String()
		resp.ProductID = &s
	}
	if oi.VariantID.Valid {
		s := uuid.UUID(oi.VariantID.Bytes).String()
		resp.VariantID = &s
	}

	return resp
}

// GetDashboardStats returns aggregate order and customer statistics for a shop.
func (s *Service) GetDashboardStats(ctx context.Context, shopID string) (*DashboardStats, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop ID")
	}
	pgShopID := pgtype.UUID{Bytes: sid, Valid: true}

	totalOrders, _ := s.q.CountOrders(ctx, sid)
	totalRevenue, _ := s.q.GetTotalRevenue(ctx, sid)
	pendingOrders, _ := s.q.CountOrdersByStatus(ctx, db.CountOrdersByStatusParams{
		ShopID: sid,
		Status: db.NullOrderStatus{OrderStatus: db.OrderStatus("pending"), Valid: true},
	})
	totalCustomers, _ := s.q.CountCustomers(ctx, pgShopID)

	return &DashboardStats{
		TotalOrders:    totalOrders,
		TotalRevenue:   pgutil.NumericToString(totalRevenue),
		PendingOrders:  pendingOrders,
		TotalCustomers: totalCustomers,
	}, nil
}

// ListRecentActivity returns the most recent audit log entries for a shop.
func (s *Service) ListRecentActivity(ctx context.Context, shopID string, limit int) ([]ActivityItem, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return []ActivityItem{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	logs, err := s.q.ListAuditLogs(ctx, db.ListAuditLogsParams{
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
		Limit:  int32(limit),
		Offset: 0,
	})
	if err != nil {
		return []ActivityItem{}, nil
	}
	items := make([]ActivityItem, len(logs))
	for i, l := range logs {
		items[i] = auditLogToActivity(l)
	}
	return items, nil
}

func auditLogToActivity(l db.AuditLog) ActivityItem {
	actorName := "System"
	if l.ActorName.Valid && l.ActorName.String != "" {
		actorName = l.ActorName.String
	}
	description := fmt.Sprintf("%s %s", l.Action, l.ResourceType)
	createdAt := ""
	if l.CreatedAt.Valid {
		createdAt = l.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	return ActivityItem{
		ID:           l.ID.String(),
		Action:       l.Action,
		ResourceType: l.ResourceType,
		ActorName:    actorName,
		Description:  description,
		CreatedAt:    createdAt,
	}
}
