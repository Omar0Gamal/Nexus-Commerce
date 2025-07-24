package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/email"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const saveLockTTL = 30 * time.Second

// SaveCart persists the authenticated customer's current Redis cart to the
// saved_carts table. It is debounced: if a save happened in the last 30 s the
// call is a no-op. Intended to be called fire-and-forget from mutation handlers.
func (s *Service) SaveCart(ctx context.Context, shopID, customerID string) error {
	if s.rdb == nil || s.q == nil {
		return nil
	}

	lockKey := fmt.Sprintf("cart:save_lock:%s:%s", shopID, customerID)
	set, err := s.rdb.SetNX(ctx, lockKey, "1", saveLockTTL).Result()
	if err != nil {
		return fmt.Errorf("acquire save lock: %w", err)
	}
	if !set {
		return nil // another save happened recently
	}

	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return fmt.Errorf("parse shop id: %w", err)
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return fmt.Errorf("parse customer id: %w", err)
	}

	identity := "customer:" + customerID
	items, err := s.store.Get(ctx, shopID, identity)
	if err != nil {
		return fmt.Errorf("get cart from redis: %w", err)
	}
	if len(items) == 0 {
		return nil
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal cart items: %w", err)
	}

	total := 0.0
	for _, it := range items {
		total += it.Price * float64(it.Quantity)
	}
	totalCents := int32(math.Round(total * 100))

	if _, err := s.q.UpsertSavedCart(ctx, db.UpsertSavedCartParams{
		ShopID:     shopUUID,
		CustomerID: customerUUID,
		Items:      raw,
		CouponCode: pgtype.Text{},
		TotalCents: totalCents,
	}); err != nil {
		return fmt.Errorf("upsert saved cart: %w", err)
	}

	return nil
}

// RecoverCart loads cart items from a one-time recovery token back into Redis,
// marks the saved_cart as recovered, and returns the hydrated Cart.
func (s *Service) RecoverCart(ctx context.Context, token string) (*Cart, string, error) {
	row, err := s.q.GetCartByRecoveryToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		return nil, "", ErrProductNotFound
	}

	var items []CartItem
	if err := json.Unmarshal(row.Items, &items); err != nil {
		return nil, "", fmt.Errorf("unmarshal saved cart: %w", err)
	}

	shopID := row.ShopID.String()
	customerID := row.CustomerID.String()
	identity := "customer:" + customerID

	if err := s.store.Save(ctx, shopID, identity, items, true); err != nil {
		return nil, "", fmt.Errorf("restore cart to redis: %w", err)
	}

	_ = s.q.MarkCartRecovered(ctx, db.MarkCartRecoveredParams{
		ShopID:     row.ShopID,
		CustomerID: row.CustomerID,
	})

	return buildCart(items), shopID, nil
}

// DeleteSavedCart removes the persisted cart row when an order is placed.
func (s *Service) DeleteSavedCart(ctx context.Context, shopID, customerID string) error {
	if s.q == nil {
		return nil
	}
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return fmt.Errorf("parse shop id: %w", err)
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return fmt.Errorf("parse customer id: %w", err)
	}
	if err := s.q.DeleteSavedCart(ctx, db.DeleteSavedCartParams{
		ShopID:     shopUUID,
		CustomerID: customerUUID,
	}); err != nil {
		return fmt.Errorf("delete saved cart: %w", err)
	}

	return nil
}

// AbandonedCartWorker scans for abandoned carts on a regular cadence and
// dispatches recovery emails for the three-touch sequence.
type AbandonedCartWorker struct {
	q      *db.Queries
	rdb    *redis.Client
	mailer *email.Mailer
	logger *zap.Logger
	// frontendURL is the base URL used to build recovery links, e.g. "https://myshop.com"
	frontendURL string
}

func NewAbandonedCartWorker(q *db.Queries, rdb *redis.Client, mailer *email.Mailer, logger *zap.Logger, frontendURL string) *AbandonedCartWorker {
	return &AbandonedCartWorker{q: q, rdb: rdb, mailer: mailer, logger: logger, frontendURL: frontendURL}
}

// Run starts the hourly processing loop. It exits when ctx is cancelled.
func (w *AbandonedCartWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *AbandonedCartWorker) process(ctx context.Context) {
	w.processEmail1(ctx)
	w.processEmail2(ctx)
	w.processEmail3(ctx)
}

func (w *AbandonedCartWorker) processEmail1(ctx context.Context) {
	carts, err := w.q.GetCartsForEmail1(ctx)
	if err != nil {
		w.logger.Error("abandoned cart email1 query", zap.Error(err))
		return
	}
	for _, c := range carts {
		skipReminder, skipErr := w.shouldSkipReminder(ctx, c.ShopID, c.CustomerID, c.UpdatedAt)
		if skipErr != nil {
			w.logger.Warn("abandoned cart email1 staleness check failed",
				zap.String("shop_id", c.ShopID.String()),
				zap.String("customer_id", c.CustomerID.String()),
				zap.Error(skipErr),
			)
			continue
		}
		if skipReminder {
			_ = w.q.MarkCartRecovered(ctx, db.MarkCartRecoveredParams{ShopID: c.ShopID, CustomerID: c.CustomerID})
			continue
		}

		token := uuid.New().String()
		_ = w.q.SetRecoveryToken(ctx, db.SetRecoveryTokenParams{
			ShopID:        c.ShopID,
			CustomerID:    c.CustomerID,
			RecoveryToken: pgtype.Text{String: token, Valid: true},
		})

		firstName := ""
		if c.CustomerFirstName.Valid {
			firstName = c.CustomerFirstName.String
		}

		items := marshalCartItemsForEmail(c.Items)
		recoverURL := fmt.Sprintf("%s/cart/recover?token=%s", w.frontendURL, token)
		if err := w.mailer.SendAbandonedCart(c.CustomerEmail, email.AbandonedCartData{
			CustomerName: firstName,
			Items:        items,
			RecoverURL:   recoverURL,
			Touch:        1,
		}); err != nil {
			w.logger.Error("abandoned cart email1 send", zap.String("to", c.CustomerEmail), zap.Error(err))
			continue
		}
		_ = w.q.MarkCartEmail1Sent(ctx, db.MarkCartEmail1SentParams{
			ShopID: c.ShopID,
			ID:     c.ID,
		})
	}
}

func (w *AbandonedCartWorker) processEmail2(ctx context.Context) {
	carts, err := w.q.GetCartsForEmail2(ctx)
	if err != nil {
		w.logger.Error("abandoned cart email2 query", zap.Error(err))
		return
	}
	for _, c := range carts {
		skipReminder, skipErr := w.shouldSkipReminder(ctx, c.ShopID, c.CustomerID, c.UpdatedAt)
		if skipErr != nil {
			w.logger.Warn("abandoned cart email2 staleness check failed",
				zap.String("shop_id", c.ShopID.String()),
				zap.String("customer_id", c.CustomerID.String()),
				zap.Error(skipErr),
			)
			continue
		}
		if skipReminder {
			_ = w.q.MarkCartRecovered(ctx, db.MarkCartRecoveredParams{ShopID: c.ShopID, CustomerID: c.CustomerID})
			continue
		}

		firstName := ""
		if c.CustomerFirstName.Valid {
			firstName = c.CustomerFirstName.String
		}

		recoverURL := ""
		if c.RecoveryToken.Valid {
			recoverURL = fmt.Sprintf("%s/cart/recover?token=%s", w.frontendURL, c.RecoveryToken.String)
		}

		items := marshalCartItemsForEmail(c.Items)
		if err := w.mailer.SendAbandonedCart(c.CustomerEmail, email.AbandonedCartData{
			CustomerName: firstName,
			Items:        items,
			RecoverURL:   recoverURL,
			Touch:        2,
		}); err != nil {
			w.logger.Error("abandoned cart email2 send", zap.String("to", c.CustomerEmail), zap.Error(err))
			continue
		}
		_ = w.q.MarkCartEmail2Sent(ctx, db.MarkCartEmail2SentParams{
			ShopID: c.ShopID,
			ID:     c.ID,
		})
	}
}

func (w *AbandonedCartWorker) processEmail3(ctx context.Context) {
	carts, err := w.q.GetCartsForEmail3(ctx)
	if err != nil {
		w.logger.Error("abandoned cart email3 query", zap.Error(err))
		return
	}
	for _, c := range carts {
		skipReminder, skipErr := w.shouldSkipReminder(ctx, c.ShopID, c.CustomerID, c.UpdatedAt)
		if skipErr != nil {
			w.logger.Warn("abandoned cart email3 staleness check failed",
				zap.String("shop_id", c.ShopID.String()),
				zap.String("customer_id", c.CustomerID.String()),
				zap.Error(skipErr),
			)
			continue
		}
		if skipReminder {
			_ = w.q.MarkCartRecovered(ctx, db.MarkCartRecoveredParams{ShopID: c.ShopID, CustomerID: c.CustomerID})
			continue
		}

		firstName := ""
		if c.CustomerFirstName.Valid {
			firstName = c.CustomerFirstName.String
		}

		recoverURL := ""
		if c.RecoveryToken.Valid {
			recoverURL = fmt.Sprintf("%s/cart/recover?token=%s", w.frontendURL, c.RecoveryToken.String)
		}

		items := marshalCartItemsForEmail(c.Items)
		if err := w.mailer.SendAbandonedCart(c.CustomerEmail, email.AbandonedCartData{
			CustomerName: firstName,
			Items:        items,
			RecoverURL:   recoverURL,
			Touch:        3,
		}); err != nil {
			w.logger.Error("abandoned cart email3 send", zap.String("to", c.CustomerEmail), zap.Error(err))
			continue
		}
		_ = w.q.MarkCartEmail3Sent(ctx, db.MarkCartEmail3SentParams{
			ShopID: c.ShopID,
			ID:     c.ID,
		})
	}
}

// marshalCartItemsForEmail decodes the JSONB cart items into the email-ready slice.
func marshalCartItemsForEmail(raw []byte) []email.AbandonedCartItem {
	var items []CartItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	result := make([]email.AbandonedCartItem, 0, len(items))
	for _, it := range items {
		result = append(result, email.AbandonedCartItem{
			Title:    it.Title,
			Quantity: it.Quantity,
			Price:    fmt.Sprintf("%.2f", it.Price),
			ImageURL: it.ImageURL,
		})
	}
	return result
}

func (w *AbandonedCartWorker) shouldSkipReminder(ctx context.Context, shopID, customerID uuid.UUID, updatedAt pgtype.Timestamptz) (bool, error) {
	if !updatedAt.Valid {
		return false, nil
	}

	hasOrder, err := w.q.HasOrderSinceSavedCart(ctx, db.HasOrderSinceSavedCartParams{
		ShopID:     shopID,
		CustomerID: pgtype.UUID{Bytes: customerID, Valid: true},
		CreatedAt:  updatedAt,
	})
	if err != nil {
		return false, err
	}

	return hasOrder, nil
}
