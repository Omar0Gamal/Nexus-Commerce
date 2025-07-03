package shipping

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"backend-api/internal/db"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotFound    = apperr.ErrNotFound
	ErrInvalidUUID = apperr.ErrInvalidUUID
	ErrNoRate      = errors.New("no shipping rate available for the destination country")
)

// Service handles shipping zone and rate management.
type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}


type CreateZoneRequest struct {
	Name    string   `json:"name" binding:"required,min=1,max=100"`
	Regions []string `json:"regions"` // list of country codes, empty = worldwide
}

type UpdateZoneRequest struct {
	Name    *string  `json:"name,omitempty"`
	Regions []string `json:"regions,omitempty"` // nil = no change
}

type ZoneResponse struct {
	ID        string   `json:"id"`
	ShopID    string   `json:"shop_id"`
	Name      string   `json:"name"`
	Regions   []string `json:"regions"`
	CreatedAt string   `json:"created_at"`
}


type CreateRateRequest struct {
	ZoneID               string   `json:"zone_id" binding:"required"`
	Name                 string   `json:"name" binding:"required,min=1,max=100"`
	RateType             string   `json:"rate_type" binding:"required,oneof=flat weight_based"`
	BaseRate             float64  `json:"base_rate" binding:"gte=0"`
	RatePerKg            *float64 `json:"rate_per_kg,omitempty"`
	MinOrderFreeShipping *float64 `json:"min_order_free_shipping,omitempty"`
	IsActive             *bool    `json:"is_active,omitempty"`
}

type UpdateRateRequest struct {
	Name                 *string  `json:"name,omitempty"`
	RateType             *string  `json:"rate_type,omitempty"`
	BaseRate             *float64 `json:"base_rate,omitempty"`
	RatePerKg            *float64 `json:"rate_per_kg,omitempty"`
	MinOrderFreeShipping *float64 `json:"min_order_free_shipping,omitempty"`
	IsActive             *bool    `json:"is_active,omitempty"`
}

type RateResponse struct {
	ID                   string  `json:"id"`
	ZoneID               string  `json:"zone_id"`
	ShopID               string  `json:"shop_id"`
	Name                 string  `json:"name"`
	RateType             string  `json:"rate_type"`
	BaseRate             string  `json:"base_rate"`
	RatePerKg            *string `json:"rate_per_kg,omitempty"`
	MinOrderFreeShipping *string `json:"min_order_free_shipping,omitempty"`
	IsActive             bool    `json:"is_active"`
	CreatedAt            string  `json:"created_at"`
}


func (s *Service) CreateZone(ctx context.Context, shopID string, req CreateZoneRequest) (*ZoneResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	regionsJSON, err := json.Marshal(req.Regions)
	if err != nil {
		return nil, fmt.Errorf("marshal regions: %w", err)
	}

	z, err := s.q.CreateShippingZone(ctx, db.CreateShippingZoneParams{
		ShopID:  shopPgUUID,
		Name:    req.Name,
		Regions: regionsJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("create zone: %w", err)
	}
	resp := mapZone(z)
	return &resp, nil
}

func (s *Service) GetZone(ctx context.Context, shopID, zoneID string) (*ZoneResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	id, err := uuid.Parse(zoneID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	z, err := s.q.GetShippingZone(ctx, db.GetShippingZoneParams{ID: id, ShopID: shopPgUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get zone: %w", err)
	}
	resp := mapZone(z)
	return &resp, nil
}

func (s *Service) ListZones(ctx context.Context, shopID string) ([]ZoneResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	rows, err := s.q.ListShippingZones(ctx, shopPgUUID)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}
	result := make([]ZoneResponse, len(rows))
	for i, z := range rows {
		result[i] = mapZone(z)
	}
	return result, nil
}

func (s *Service) UpdateZone(ctx context.Context, shopID, zoneID string, req UpdateZoneRequest) (*ZoneResponse, error) {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	id, err := uuid.Parse(zoneID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	params := db.UpdateShippingZoneParams{ID: id, ShopID: shopPgUUID}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Regions != nil {
		b, _ := json.Marshal(req.Regions)
		params.Regions = b
	}
	z, err := s.q.UpdateShippingZone(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update zone: %w", err)
	}
	resp := mapZone(z)
	return &resp, nil
}

func (s *Service) DeleteZone(ctx context.Context, shopID, zoneID string) error {
	shopPgUUID, err := pgutil.ParseUUID(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	id, err := uuid.Parse(zoneID)
	if err != nil {
		return ErrInvalidUUID
	}
	return s.q.DeleteShippingZone(ctx, db.DeleteShippingZoneParams{ID: id, ShopID: shopPgUUID})
}


func (s *Service) CreateRate(ctx context.Context, shopID string, req CreateRateRequest) (*RateResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	zoneUUID, err := uuid.Parse(req.ZoneID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	params := db.CreateShippingRateParams{
		ZoneID:   zoneUUID,
		ShopID:   shopUUID,
		Name:     req.Name,
		RateType: db.ShippingRateType(strings.ToLower(req.RateType)),
		BaseRate: pgutil.FloatToNumericCents(req.BaseRate),
		IsActive: isActive,
	}
	if req.RatePerKg != nil {
		params.RatePerKg = pgutil.FloatToNumericCents(*req.RatePerKg)
	}
	if req.MinOrderFreeShipping != nil {
		params.MinOrderFreeShipping = pgutil.FloatToNumericCents(*req.MinOrderFreeShipping)
	}

	r, err := s.q.CreateShippingRate(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create rate: %w", err)
	}
	resp := mapRate(r)
	return &resp, nil
}

func (s *Service) GetRate(ctx context.Context, shopID, rateID string) (*RateResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	id, err := uuid.Parse(rateID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	r, err := s.q.GetShippingRate(ctx, db.GetShippingRateParams{ID: id, ShopID: shopUUID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get rate: %w", err)
	}
	resp := mapRate(r)
	return &resp, nil
}

func (s *Service) ListRates(ctx context.Context, shopID string) ([]RateResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	rows, err := s.q.ListShippingRates(ctx, shopUUID)
	if err != nil {
		return nil, fmt.Errorf("list rates: %w", err)
	}
	result := make([]RateResponse, len(rows))
	for i, r := range rows {
		result[i] = mapRate(r)
	}
	return result, nil
}

func (s *Service) ListRatesByZone(ctx context.Context, shopID, zoneID string) ([]RateResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	zoneUUID, err := uuid.Parse(zoneID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	rows, err := s.q.ListRatesByZone(ctx, db.ListRatesByZoneParams{ZoneID: zoneUUID, ShopID: shopUUID})
	if err != nil {
		return nil, fmt.Errorf("list rates by zone: %w", err)
	}
	result := make([]RateResponse, len(rows))
	for i, r := range rows {
		result[i] = mapRate(r)
	}
	return result, nil
}

func (s *Service) DeleteRate(ctx context.Context, shopID, rateID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	id, err := uuid.Parse(rateID)
	if err != nil {
		return ErrInvalidUUID
	}
	return s.q.DeleteShippingRate(ctx, db.DeleteShippingRateParams{ID: id, ShopID: shopUUID})
}

func (s *Service) UpdateRate(ctx context.Context, shopID, rateID string, req UpdateRateRequest) (*RateResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	id, err := uuid.Parse(rateID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	params := db.UpdateShippingRateParams{ID: id, ShopID: shopUUID}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.RateType != nil {
		params.RateType = db.NullShippingRateType{ShippingRateType: db.ShippingRateType(*req.RateType), Valid: true}
	}
	if req.BaseRate != nil {
		params.BaseRate = pgutil.FloatToNumericCents(*req.BaseRate)
	}
	if req.RatePerKg != nil {
		params.RatePerKg = pgutil.FloatToNumericCents(*req.RatePerKg)
	}
	if req.MinOrderFreeShipping != nil {
		params.MinOrderFreeShipping = pgutil.FloatToNumericCents(*req.MinOrderFreeShipping)
	}
	if req.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *req.IsActive, Valid: true}
	}
	r, err := s.q.UpdateShippingRate(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update rate: %w", err)
	}
	resp := mapRate(r)
	return &resp, nil
}


// ResolveShippingFee returns the shipping fee for a destination country and
// order total. If a matching rate is set to free-above-threshold and the order
// total qualifies, returns 0.  If no rate matches, returns ErrNoRate.
// weightKg is optional (pass 0 to use flat rates only).
func (s *Service) ResolveShippingFee(ctx context.Context, shopID, countryCode string, orderTotal, weightKg float64) (float64, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return 0, ErrInvalidUUID
	}
	shopPgUUID := pgtype.UUID{Bytes: shopUUID, Valid: true}

	code := strings.ToUpper(strings.TrimSpace(countryCode))
	rate, err := s.q.FindRateByCountry(ctx, db.FindRateByCountryParams{
		ShopID:  shopPgUUID,
		Column2: code,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNoRate
		}
		return 0, fmt.Errorf("find shipping rate: %w", err)
	}

	// Check if order total qualifies for free shipping
	if rate.MinOrderFreeShipping.Valid {
		mofs, _ := rate.MinOrderFreeShipping.Float64Value()
		if mofs.Valid && orderTotal >= mofs.Float64 {
			return 0, nil
		}
	}

	baseRate, _ := rate.BaseRate.Float64Value()
	fee := baseRate.Float64

	if rate.RateType == db.ShippingRateTypeWeightBased && weightKg > 0 && rate.RatePerKg.Valid {
		ratePerKg, _ := rate.RatePerKg.Float64Value()
		if ratePerKg.Valid {
			fee += math.Round(ratePerKg.Float64*weightKg*100) / 100
		}
	}

	return fee, nil
}


func mapZone(z db.ShippingZone) ZoneResponse {
	resp := ZoneResponse{
		ID:      z.ID.String(),
		ShopID:  uuid.UUID(z.ShopID.Bytes).String(),
		Name:    z.Name,
		Regions: []string{},
	}
	if z.CreatedAt.Valid {
		resp.CreatedAt = z.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	// Unmarshal regions JSON
	if len(z.Regions) > 0 {
		var regions []string
		if err := json.Unmarshal(z.Regions, &regions); err == nil {
			resp.Regions = regions
		}
	}
	return resp
}

func mapRate(r db.ShippingRate) RateResponse {
	resp := RateResponse{
		ID:       r.ID.String(),
		ZoneID:   r.ZoneID.String(),
		ShopID:   r.ShopID.String(),
		Name:     r.Name,
		RateType: string(r.RateType),
		BaseRate: pgutil.NumericToString(r.BaseRate),
		IsActive: r.IsActive,
	}
	if r.CreatedAt.Valid {
		resp.CreatedAt = r.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	if r.RatePerKg.Valid {
		s := pgutil.NumericToString(r.RatePerKg)
		resp.RatePerKg = &s
	}
	if r.MinOrderFreeShipping.Valid {
		s := pgutil.NumericToString(r.MinOrderFreeShipping)
		resp.MinOrderFreeShipping = &s
	}
	return resp
}
