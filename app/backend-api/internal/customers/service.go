package customers

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"backend-api/internal/db"
	"backend-api/internal/orders"
	"backend-api/internal/worker"
	"backend-api/internal/shared/apperr"
	"backend-api/internal/shared/pgutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = apperr.ErrNotFound

type Service struct {
	q        *db.Queries
	pool     *pgxpool.Pool
	orderSvc *orders.Service
	queue    *worker.Queue
}

func NewService(q *db.Queries, pool *pgxpool.Pool, orderSvc *orders.Service, queue *worker.Queue) *Service {
	return &Service{q: q, pool: pool, orderSvc: orderSvc, queue: queue}
}

func (s *Service) ListCustomers(ctx context.Context, shopID string, params ListCustomersParams) ([]CustomerSummary, int64, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, 0, errors.New("invalid shop ID")
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}
	offset := int32((params.Page - 1) * params.PerPage)

	if s.pool != nil {
		query := `
			SELECT 
				c.id, c.email, COALESCE(c.first_name, ''), COALESCE(c.last_name, ''), COALESCE(c.phone, ''),
				COALESCE(TO_CHAR(c.created_at, 'YYYY-MM-DD"T"HH24:MI:SSZ'), ''),
				COUNT(o.id) as total_orders,
				COALESCE(SUM(CASE WHEN o.status NOT IN ('cancelled', 'refunded') THEN o.total_price ELSE 0 END), 0) as total_spent,
				COALESCE(TO_CHAR(MAX(o.created_at), 'YYYY-MM-DD'), 'Never') as last_order_date
			FROM customers c
			LEFT JOIN orders o ON o.customer_id = c.id AND o.shop_id = c.shop_id
			WHERE c.shop_id = $1 AND c.deleted_at IS NULL
			  AND ($4 = '' OR c.email ILIKE '%' || $4 || '%' OR c.first_name ILIKE '%' || $4 || '%' OR c.last_name ILIKE '%' || $4 || '%')
			GROUP BY c.id, c.email, c.first_name, c.last_name, c.phone, c.created_at
			ORDER BY total_spent DESC, c.created_at DESC
			LIMIT $2 OFFSET $3
		`
		var countQuery string
		var countArgs []interface{}
		if params.Search != "" {
			countQuery = `SELECT COUNT(*) FROM customers WHERE shop_id = $1 AND deleted_at IS NULL AND (email ILIKE '%' || $2 || '%' OR first_name ILIKE '%' || $2 || '%' OR last_name ILIKE '%' || $2 || '%')`
			countArgs = []interface{}{sid, params.Search}
		} else {
			countQuery = `SELECT COUNT(*) FROM customers WHERE shop_id = $1 AND deleted_at IS NULL`
			countArgs = []interface{}{sid}
		}
		var total int64
		_ = s.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)

		rows, err := s.pool.Query(ctx, query, sid, params.PerPage, offset, params.Search)
		if err != nil {
			return nil, 0, fmt.Errorf("list customers: %w", err)
		}
		defer rows.Close()

		var summaries []CustomerSummary
		for rows.Next() {
			var cs CustomerSummary
			var cid uuid.UUID
			var spent float64
			if err := rows.Scan(&cid, &cs.Email, &cs.FirstName, &cs.LastName, &cs.Phone, &cs.CreatedAt, &cs.TotalOrders, &spent, &cs.LastOrderDate); err == nil {
				cs.ID = cid.String()
				cs.TotalSpent = fmt.Sprintf("%.2f", spent)
				summaries = append(summaries, cs)
			}
		}
		return summaries, total, nil
	}

	pgShopID := pgtype.UUID{Bytes: sid, Valid: true}
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
	customerSummary := toSummary(c)
	return &customerSummary, nil
}

func (s *Service) ListArchivedCustomers(ctx context.Context, shopID string) ([]CustomerSummary, error) {
	sid, err := uuid.Parse(shopID)
	if err != nil {
		return nil, errors.New("invalid shop ID")
	}
	rows, err := s.q.ListDeletedCustomers(ctx, pgtype.UUID{Bytes: sid, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list archived customers: %w", err)
	}
	summaries := make([]CustomerSummary, len(rows))
	for i, c := range rows {
		summaries[i] = toSummary(c)
	}
	return summaries, nil
}

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

type CustomerDataExport struct {
	Customer CustomerSummary `json:"customer"`
	Orders   []any           `json:"orders"`
	Consents []db.ConsentLog `json:"consents"`
}

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


func (s *Service) SendWinbackEmail(ctx context.Context, shopID, customerID string) error {
	c, err := s.GetCustomer(ctx, shopID, customerID)
	if err != nil {
		return err
	}
	
	couponCode := "WINBACK-" + customerID[:8]
	
	payload := map[string]any{
		"to": c.Email,
		"coupon_code": couponCode,
	}
	
	if err := s.queue.EnqueueTyped(ctx, "email:winback", payload); err != nil {
		return err
	}
	return nil
}
