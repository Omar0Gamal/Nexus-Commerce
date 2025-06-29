package customers

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"backend-api/internal/db"
	"backend-api/internal/orders"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrNotFound is returned when a customer cannot be found.
var ErrNotFound = apperr.ErrNotFound

// Service handles customer management for shop staff.
type Service struct {
	q        *db.Queries
	orderSvc *orders.Service
}

func NewService(q *db.Queries, orderSvc *orders.Service) *Service {
	return &Service{q: q, orderSvc: orderSvc}
}

// ListCustomers returns a paginated list of customers, optionally filtered by search query.
func (s *Service) ListCustomers(ctx context.Context, shopID string, params ListCustomersParams) ([]CustomerSummary, int64, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, 0, errors.New("invalid shop ID")
	}
	pgShopID := pgtype.UUID{Bytes: sid, Valid: true}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}
	offset := int32((params.Page - 1) * params.PerPage)

	var rows []db.Customer
	if params.Search != "" {
		rows, err = s.q.SearchCustomers(ctx, db.SearchCustomersParams{
			ShopID:  pgShopID,
			Column2: pgtype.Text{String: params.Search, Valid: true},
			Limit:   int32(params.PerPage),
			Offset:  offset,
		})
	} else {
		rows, err = s.q.ListCustomers(ctx, db.ListCustomersParams{
			ShopID: pgShopID,
			Limit:  int32(params.PerPage),
			Offset: offset,
		})
	}
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}

	total, err := s.q.CountCustomers(ctx, pgShopID)
	if err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	summaries := make([]CustomerSummary, len(rows))
	for i, c := range rows {
		summaries[i] = toSummary(c)
	}
	return summaries, total, nil
}

// GetCustomer returns a single customer with their full order history.
func (s *Service) GetCustomer(ctx context.Context, shopID, customerID string) (*CustomerDetail, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrNotFound
	}
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrNotFound
	}

	c, err := s.q.GetCustomer(ctx, db.GetCustomerParams{
		ID:     cid,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}

	// Use a single aggregate query instead of loading all orders.
	summary, err := s.q.GetCustomerOrderSummary(ctx, db.GetCustomerOrderSummaryParams{
		ShopID:     sid,
		CustomerID: pgtype.UUID{Bytes: cid, Valid: true},
	})
	if err != nil {
		summary = db.GetCustomerOrderSummaryRow{}
	}

	return &CustomerDetail{
		CustomerSummary: toSummary(c),
		TotalOrders:     int(summary.OrderCount),
		TotalSpent:      pgutil.NumericToString(summary.TotalSpent),
		Orders:          []orders.OrderResponse{},
	}, nil
}

// SoftDeleteCustomer moves a customer to the archive (sets deleted_at).
func (s *Service) SoftDeleteCustomer(ctx context.Context, shopID, customerID string) error {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return ErrNotFound
	}
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return ErrNotFound
	}
	return s.q.SoftDeleteCustomer(ctx, db.SoftDeleteCustomerParams{
		ID:     cid,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
}

// RestoreCustomer un-archives a previously soft-deleted customer.
func (s *Service) RestoreCustomer(ctx context.Context, shopID, customerID string) (*CustomerSummary, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrNotFound
	}
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrNotFound
	}
	c, err := s.q.RestoreCustomer(ctx, db.RestoreCustomerParams{
		ID:     cid,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("restore customer: %w", err)
	}
	result := toSummary(c)
	return &result, nil
}

// ListArchivedCustomers returns all soft-deleted customers for a shop.
func (s *Service) ListArchivedCustomers(ctx context.Context, shopID string) ([]CustomerSummary, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, errors.New("invalid shop ID")
	}
	rows, err := s.q.ListDeletedCustomers(ctx, pgtype.UUID{Bytes: sid, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list archived customers: %w", err)
	}
	result := make([]CustomerSummary, len(rows))
	for i, c := range rows {
		result[i] = toSummary(c)
	}
	return result, nil
}



// LogConsent records a customer's consent action (e.g. "terms_v2", "marketing").
func (s *Service) LogConsent(ctx context.Context, customerID, shopID, consentType, ipAddress, userAgent, termsVersion string) error {
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return fmt.Errorf("invalid customer ID: %w", err)
	}
	var ipAddr *netip.Addr
	if ipAddress != "" {
		parsed, err := netip.ParseAddr(ipAddress)
		if err == nil {
			ipAddr = &parsed
		}
	}
	_, err = s.q.CreateConsentLog(ctx, db.CreateConsentLogParams{
		CustomerID:   cid,
		ConsentType:  consentType,
		IpAddress:    ipAddr,
		UserAgent:    pgtype.Text{String: userAgent, Valid: userAgent != ""},
		TermsVersion: pgtype.Text{String: termsVersion, Valid: termsVersion != ""},
	})
	return err
}

// CustomerDataExport holds the exported customer data for GDPR.
type CustomerDataExport struct {
	Customer CustomerSummary `json:"customer"`
	Orders   []any           `json:"orders"`
	Consents []db.ConsentLog `json:"consents"`
}

// ExportCustomerData assembles all personal data for the given customer.
func (s *Service) ExportCustomerData(ctx context.Context, customerID, shopID string) (*CustomerDataExport, error) {
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, fmt.Errorf("invalid shop ID: %w", err)
	}

	customer, err := s.q.GetCustomer(ctx, db.GetCustomerParams{
		ID:     cid,
		ShopID: pgtype.UUID{Bytes: sid, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}

	consents, _ := s.q.ListConsentsByCustomer(ctx, cid)

	orders, _ := s.q.ListOrdersByCustomer(ctx, db.ListOrdersByCustomerParams{
		CustomerID: pgtype.UUID{Bytes: cid, Valid: true},
		ShopID:     sid,
	})

	ordersAny := make([]any, len(orders))
	for i, o := range orders {
		ordersAny[i] = o
	}

	return &CustomerDataExport{
		Customer: toSummary(customer),
		Orders:   ordersAny,
		Consents: consents,
	}, nil
}

// AnonymizeCustomer anonymises the customer PII then soft-deletes the record.
func (s *Service) AnonymizeCustomer(ctx context.Context, customerID, shopID string) error {
	cid, err := uuid.Parse(customerID)
	if err != nil {
		return fmt.Errorf("invalid customer ID: %w", err)
	}
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return fmt.Errorf("invalid shop ID: %w", err)
	}
	pgShopID := pgtype.UUID{Bytes: sid, Valid: true}

	anonEmail := fmt.Sprintf("deleted+%s@example.com", cid.String())
	if _, err := s.q.UpdateCustomer(ctx, db.UpdateCustomerParams{
		ID:        cid,
		ShopID:    pgShopID,
		Email:     pgtype.Text{String: anonEmail, Valid: true},
		FirstName: pgtype.Text{String: "Deleted", Valid: true},
		LastName:  pgtype.Text{String: "User", Valid: true},
		Phone:     pgtype.Text{String: "", Valid: true},
	}); err != nil {
		return fmt.Errorf("anonymize customer: %w", err)
	}

	return s.q.SoftDeleteCustomer(ctx, db.SoftDeleteCustomerParams{
		ID:     cid,
		ShopID: pgShopID,
	})
}

func toSummary(c db.Customer) CustomerSummary {
	s := CustomerSummary{
		ID:    c.ID.String(),
		Email: c.Email,
	}
	if c.FirstName.Valid {
		s.FirstName = c.FirstName.String
	}
	if c.LastName.Valid {
		s.LastName = c.LastName.String
	}
	if c.Phone.Valid {
		s.Phone = c.Phone.String
	}
	if c.CreatedAt.Valid {
		s.CreatedAt = c.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	return s
}
