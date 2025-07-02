package coupons

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"backend-api/internal/db"
	sharedanalytics "backend-api/internal/shared/analytics"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound          = apperr.ErrNotFound
	ErrInvalidUUID       = apperr.ErrInvalidUUID
	ErrInvalidType       = errors.New("invalid discount type: must be percentage or fixed_amount")
	ErrCodeTaken         = errors.New("coupon code already exists for this shop")
	ErrExpired           = errors.New("coupon has expired")
	ErrUsageLimitReached = errors.New("coupon usage limit has been reached")
	ErrMinOrder          = errors.New("order total does not meet the minimum required for this coupon")
	ErrInactive          = errors.New("coupon is not active")
)

var validTypes = map[string]bool{
	"percentage":   true,
	"fixed_amount": true,
}

// Service handles coupon business logic.
type Service struct {
	q   *db.Queries
	rdb *redis.Client
}

func NewService(q *db.Queries, rdb *redis.Client) *Service {
	return &Service{q: q, rdb: rdb}
}


type CreateCouponRequest struct {
	Code           string   `json:"code" binding:"required,min=1,max=50"`
	Type           string   `json:"type" binding:"required"`
	Value          float64  `json:"value" binding:"required,gt=0"`
	IsActive       *bool    `json:"is_active,omitempty"`
	ExpiresAt      *string  `json:"expires_at,omitempty"` // RFC3339
	MinOrderAmount *float64 `json:"min_order_amount,omitempty"`
	UsageLimit     *int32   `json:"usage_limit,omitempty"`
}

type UpdateCouponRequest struct {
	Code           *string  `json:"code,omitempty"`
	Type           *string  `json:"type,omitempty"`
	Value          *float64 `json:"value,omitempty"`
	IsActive       *bool    `json:"is_active,omitempty"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
	MinOrderAmount *float64 `json:"min_order_amount,omitempty"`
	UsageLimit     *int32   `json:"usage_limit,omitempty"`
}

type CouponResponse struct {
	ID             string  `json:"id"`
	ShopID         string  `json:"shop_id"`
	Code           string  `json:"code"`
	Type           string  `json:"type"`
	Value          string  `json:"value"`
	IsActive       bool    `json:"is_active"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	MinOrderAmount *string `json:"min_order_amount,omitempty"`
	UsageLimit     *int32  `json:"usage_limit,omitempty"`
	UsageCount     int32   `json:"usage_count"`
}

// ApplyResult is returned from ApplyCoupon for use in CreateOrder.
type ApplyResult struct {
	CouponID       uuid.UUID
	DiscountAmount float64 // computed discount in same currency as order total
}


func (s *Service) CreateCoupon(ctx context.Context, shopID string, req CreateCouponRequest) (*CouponResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	dt := strings.ToLower(req.Type)
	if !validTypes[dt] {
		return nil, ErrInvalidType
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	params := db.CreateCouponParams{
		ShopID:   shopPgUUID,
		Code:     strings.ToUpper(strings.TrimSpace(req.Code)),
		Type:     db.DiscountType(dt),
		Value:    pgutil.FloatToNumericCents(req.Value),
		IsActive: pgtype.Bool{Bool: isActive, Valid: true},
	}

	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("expires_at must be RFC3339: %w", err)
		}
		params.ExpiresAt = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if req.MinOrderAmount != nil {
		params.MinOrderAmount = pgutil.FloatToNumericCents(*req.MinOrderAmount)
	}
	if req.UsageLimit != nil {
		params.UsageLimit = pgtype.Int4{Int32: *req.UsageLimit, Valid: true}
	}

	c, err := s.q.CreateCoupon(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		return nil, fmt.Errorf("create coupon: %w", err)
	}
	resp := mapCoupon(c)
	return &resp, nil
}

func (s *Service) GetCoupon(ctx context.Context, shopID, couponID string) (*CouponResponse, error) {
	id, err := uuid.Parse(couponID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	c, err := s.q.GetCoupon(ctx, db.GetCouponParams{ID: id, ShopID: shopPgUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get coupon: %w", err)
	}
	resp := mapCoupon(c)
	return &resp, nil
}

func (s *Service) ListCoupons(ctx context.Context, shopID string) ([]CouponResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	rows, err := s.q.ListCoupons(ctx, shopPgUUID)
	if err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	result := make([]CouponResponse, len(rows))
	for i, r := range rows {
		result[i] = mapCoupon(r)
	}
	return result, nil
}

func (s *Service) UpdateCoupon(ctx context.Context, shopID, couponID string, req UpdateCouponRequest) (*CouponResponse, error) {
	id, err := uuid.Parse(couponID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	if _, err := s.q.GetCoupon(ctx, db.GetCouponParams{ID: id, ShopID: shopPgUUID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get coupon: %w", err)
	}

	params := db.UpdateCouponParams{ID: id, ShopID: shopPgUUID}
	if req.Code != nil {
		params.Code = pgtype.Text{String: strings.ToUpper(strings.TrimSpace(*req.Code)), Valid: true}
	}
	if req.Type != nil {
		dt := strings.ToLower(*req.Type)
		if !validTypes[dt] {
			return nil, ErrInvalidType
		}
		params.Type = db.NullDiscountType{DiscountType: db.DiscountType(dt), Valid: true}
	}
	if req.Value != nil {
		params.Value = pgtype.Numeric{Int: big.NewInt(int64(math.Round(*req.Value * 100))), Exp: -2, Valid: true}
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("expires_at must be RFC3339: %w", err)
		}
		params.ExpiresAt = pgtype.Timestamptz{Time: t, Valid: true}
	}
	if req.MinOrderAmount != nil {
		params.MinOrderAmount = pgtype.Numeric{Int: big.NewInt(int64(math.Round(*req.MinOrderAmount * 100))), Exp: -2, Valid: true}
	}
	if req.UsageLimit != nil {
		params.UsageLimit = pgtype.Int4{Int32: *req.UsageLimit, Valid: true}
	}

	c, err := s.q.UpdateCoupon(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update coupon: %w", err)
	}
	resp := mapCoupon(c)
	return &resp, nil
}

func (s *Service) ToggleCoupon(ctx context.Context, shopID, couponID string) (*CouponResponse, error) {
	id, err := uuid.Parse(couponID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	c, err := s.q.ToggleCoupon(ctx, db.ToggleCouponParams{ID: id, ShopID: shopPgUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("toggle coupon: %w", err)
	}
	resp := mapCoupon(c)
	return &resp, nil
}

func (s *Service) DeleteCoupon(ctx context.Context, shopID, couponID string) error {
	id, err := uuid.Parse(couponID)
	if err != nil {
		return ErrInvalidUUID
	}
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	return s.q.DeleteCoupon(ctx, db.DeleteCouponParams{ID: id, ShopID: shopPgUUID})
}


// ApplyCoupon validates the coupon code against the given order total (in the
// shop's currency) and returns the discount to apply. The caller is responsible
// for calling IncrementUsage after the order is committed.
func (s *Service) ApplyCoupon(ctx context.Context, shopID, code string, orderTotal float64) (*ApplyResult, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	c, err := s.q.ValidateCoupon(ctx, db.ValidateCouponParams{
		Code:   strings.ToUpper(strings.TrimSpace(code)),
		ShopID: shopPgUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("validate coupon: %w", err)
	}

	// Min order amount check
	if c.MinOrderAmount.Valid {
		minF, _ := c.MinOrderAmount.Float64Value()
		if minF.Valid && orderTotal < minF.Float64 {
			return nil, ErrMinOrder
		}
	}

	couponValue, _ := c.Value.Float64Value()
	var discount float64
	switch c.Type {
	case db.DiscountTypePercentage:
		pct := couponValue.Float64
		if pct > 100 {
			pct = 100
		}
		discount = math.Round(orderTotal*pct/100*100) / 100
	case db.DiscountTypeFixedAmount:
		discount = couponValue.Float64
		if discount > orderTotal {
			discount = orderTotal
		}
	}

	result := &ApplyResult{
		CouponID:       c.ID,
		DiscountAmount: discount,
	}

	// Publish analytics event (fire-and-forget)
	if s.rdb != nil {
		discountCents := int64(discount * 100)
		sharedanalytics.Publish(ctx, s.rdb, shopID, sharedanalytics.Event{
			Event:  "coupon_applied",
			ShopID: shopID,
			Payload: map[string]any{
				"coupon_code":    strings.ToUpper(strings.TrimSpace(code)),
				"discount_cents": discountCents,
			},
		})
	}

	return result, nil
}

// IncrementUsage bumps the coupon's usage_count. Call after order commit.
func (s *Service) IncrementUsage(ctx context.Context, shopID string, couponID uuid.UUID) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return
	}
	_, _ = s.q.IncrementCouponUsage(ctx, db.IncrementCouponUsageParams{
		ID:     couponID,
		ShopID: shopPgUUID,
	})
}


func mapCoupon(c db.Coupon) CouponResponse {
	resp := CouponResponse{
		ID:         c.ID.String(),
		ShopID:     uuid.UUID(c.ShopID.Bytes).String(),
		Code:       c.Code,
		Type:       string(c.Type),
		Value:      pgutil.NumericToString(c.Value),
		IsActive:   c.IsActive.Bool,
		UsageCount: c.UsageCount,
	}
	if c.ExpiresAt.Valid {
		s := c.ExpiresAt.Time.Format(time.RFC3339)
		resp.ExpiresAt = &s
	}
	if c.MinOrderAmount.Valid {
		s := pgutil.NumericToString(c.MinOrderAmount)
		resp.MinOrderAmount = &s
	}
	if c.UsageLimit.Valid {
		resp.UsageLimit = &c.UsageLimit.Int32
	}
	return resp
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
