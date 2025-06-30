package addresses

import (
	"context"
	"errors"
	"fmt"

	"backend-api/internal/db"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when an address cannot be found.
var ErrNotFound = apperr.ErrNotFound

// Service handles customer address management.
type Service struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewService(q *db.Queries, pool *pgxpool.Pool) *Service {
	return &Service{q: q, pool: pool}
}

// ListAddresses returns all saved addresses for a customer.
func (s *Service) ListAddresses(ctx context.Context, shopID, customerID string) ([]AddressResponse, error) {
	sid, cid, err := parseIDs(shopID, customerID)
	if err != nil {
		return nil, err
	}

	rows, err := s.q.ListCustomerAddresses(ctx, db.ListCustomerAddressesParams{
		CustomerID: cid,
		ShopID:     sid,
	})
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}

	result := make([]AddressResponse, len(rows))
	for i, row := range rows {
		result[i] = toResponse(row)
	}
	return result, nil
}

// CreateAddress saves a new address for a customer.
func (s *Service) CreateAddress(ctx context.Context, shopID, customerID string, req CreateAddressRequest) (*AddressResponse, error) {
	sid, cid, err := parseIDs(shopID, customerID)
	if err != nil {
		return nil, err
	}

	label := req.Label
	if label == "" {
		label = "Home"
	}
	country := req.Country
	if country == "" {
		country = "Egypt"
	}

	// If this should be default, clear existing defaults first.
	if req.IsDefault {
		_ = s.q.ClearDefaultAddresses(ctx, db.ClearDefaultAddressesParams{
			CustomerID: cid,
			ShopID:     sid,
		})
	}

	addr, err := s.q.CreateCustomerAddress(ctx, db.CreateCustomerAddressParams{
		CustomerID: cid,
		ShopID:     sid,
		Label:      label,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Address1:   req.Address1,
		Address2:   pgutil.ToText(req.Address2),
		City:       req.City,
		State:      pgutil.ToText(req.State),
		Country:    country,
		ZipCode:    pgutil.ToText(req.ZipCode),
		Phone:      pgutil.ToText(req.Phone),
		IsDefault:  req.IsDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("create address: %w", err)
	}

	resp := toResponse(addr)
	return &resp, nil
}

// UpdateAddress modifies an existing address.
func (s *Service) UpdateAddress(ctx context.Context, shopID, customerID, addressID string, req UpdateAddressRequest) (*AddressResponse, error) {
	sid, cid, err := parseIDs(shopID, customerID)
	if err != nil {
		return nil, err
	}
	aid, err := uuid.Parse(addressID)
	if err != nil {
		return nil, ErrNotFound
	}

	addr, err := s.q.UpdateCustomerAddress(ctx, db.UpdateCustomerAddressParams{
		ID:         aid,
		CustomerID: cid,
		ShopID:     sid,
		Label:      pgutil.ToText(req.Label),
		FirstName:  pgutil.ToText(req.FirstName),
		LastName:   pgutil.ToText(req.LastName),
		Address1:   pgutil.ToText(req.Address1),
		Address2:   pgutil.ToText(req.Address2),
		City:       pgutil.ToText(req.City),
		State:      pgutil.ToText(req.State),
		Country:    pgutil.ToText(req.Country),
		ZipCode:    pgutil.ToText(req.ZipCode),
		Phone:      pgutil.ToText(req.Phone),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update address: %w", err)
	}

	resp := toResponse(addr)
	return &resp, nil
}

// DeleteAddress removes an address.
func (s *Service) DeleteAddress(ctx context.Context, shopID, customerID, addressID string) error {
	sid, cid, err := parseIDs(shopID, customerID)
	if err != nil {
		return ErrNotFound
	}
	aid, err := uuid.Parse(addressID)
	if err != nil {
		return ErrNotFound
	}

	return s.q.DeleteCustomerAddress(ctx, db.DeleteCustomerAddressParams{
		ID:         aid,
		CustomerID: cid,
		ShopID:     sid,
	})
}

// SetDefault marks an address as the default and clears others.
func (s *Service) SetDefault(ctx context.Context, shopID, customerID, addressID string) error {
	sid, cid, err := parseIDs(shopID, customerID)
	if err != nil {
		return ErrNotFound
	}
	aid, err := uuid.Parse(addressID)
	if err != nil {
		return ErrNotFound
	}

	// Verify address belongs to this customer before making it default.
	_, err = s.q.GetCustomerAddress(ctx, db.GetCustomerAddressParams{
		ID:         aid,
		CustomerID: cid,
		ShopID:     sid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get address: %w", err)
	}

	// Wrap in a transaction so we never leave the customer without a default.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := s.q.WithTx(tx)

	if err := qtx.ClearDefaultAddresses(ctx, db.ClearDefaultAddressesParams{
		CustomerID: cid,
		ShopID:     sid,
	}); err != nil {
		return fmt.Errorf("clear defaults: %w", err)
	}

	if err := qtx.SetAddressAsDefault(ctx, db.SetAddressAsDefaultParams{
		ID:         aid,
		CustomerID: cid,
		ShopID:     sid,
	}); err != nil {
		return fmt.Errorf("set default: %w", err)
	}

	return tx.Commit(ctx)
}

func parseIDs(shopID, customerID string) (uuid.UUID, uuid.UUID, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New("invalid shop ID")
	}
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, errors.New("invalid customer ID")
	}
	return sid, cid, nil
}

func toResponse(a db.CustomerAddress) AddressResponse {
	resp := AddressResponse{
		ID:        a.ID.String(),
		Label:     a.Label,
		FirstName: a.FirstName,
		LastName:  a.LastName,
		Address1:  a.Address1,
		City:      a.City,
		Country:   a.Country,
		IsDefault: a.IsDefault,
	}
	if a.Address2.Valid {
		resp.Address2 = a.Address2.String
	}
	if a.State.Valid {
		resp.State = a.State.String
	}
	if a.ZipCode.Valid {
		resp.ZipCode = a.ZipCode.String
	}
	if a.Phone.Valid {
		resp.Phone = a.Phone.String
	}
	if a.CreatedAt.Valid {
		resp.CreatedAt = a.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	return resp
}
