package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

var (
	ErrPlanNotFound = errors.New("plan not found")
	ErrShopNotFound = errors.New("shop not found")
	ErrProductLimit = errors.New("product limit reached for current plan")
	ErrStaffLimit   = errors.New("staff seat limit reached for current plan")
	ErrInvalidPlan  = errors.New("invalid plan ID")
)

type Service struct {
	q   *db.Queries
	rdb *redis.Client
}

func NewService(q *db.Queries, rdb *redis.Client) *Service {
	return &Service{q: q, rdb: rdb}
}

func (s *Service) Q() *db.Queries { return s.q }

type PlanResponse struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	MonthlyPrice          string          `json:"monthly_price"`
	MaxProducts           *int32          `json:"max_products"`
	MaxStaffAccounts      *int32          `json:"max_staff_accounts"`
	MaxStorageMb          *int32          `json:"max_storage_mb"`
	TransactionFeePercent string          `json:"transaction_fee_percent"`
	Features              json.RawMessage `json:"features"`
	CreatedAt             string          `json:"created_at"`
}

func planToResponse(p db.Plan) PlanResponse {
	var maxProducts *int32
	if p.MaxProducts.Valid {
		v := p.MaxProducts.Int32
		maxProducts = &v
	}
	var maxStaff *int32
	if p.MaxStaffAccounts.Valid {
		v := p.MaxStaffAccounts.Int32
		maxStaff = &v
	}
	var maxStorage *int32
	if p.MaxStorageMb.Valid {
		v := p.MaxStorageMb.Int32
		maxStorage = &v
	}

	monthlyPrice := ""
	if p.MonthlyPrice.Valid {
		monthlyPrice = p.MonthlyPrice.Int.String()
		if p.MonthlyPrice.Exp != 0 {
			monthlyPrice = fmt.Sprintf("%se%d", monthlyPrice, p.MonthlyPrice.Exp)
		}
	}

	txFee := ""
	if p.TransactionFeePercent.Valid {
		txFee = p.TransactionFeePercent.Int.String()
		if p.TransactionFeePercent.Exp != 0 {
			txFee = fmt.Sprintf("%se%d", txFee, p.TransactionFeePercent.Exp)
		}
	}

	var features json.RawMessage
	if len(p.Features) > 0 {
		features = json.RawMessage(p.Features)
	} else {
		features = json.RawMessage(`{}`)
	}

	return PlanResponse{
		ID:                    p.ID.String(),
		Name:                  p.Name,
		MonthlyPrice:          monthlyPrice,
		MaxProducts:           maxProducts,
		MaxStaffAccounts:      maxStaff,
		MaxStorageMb:          maxStorage,
		TransactionFeePercent: txFee,
		Features:              features,
		CreatedAt:             p.CreatedAt.Time.Format(time.RFC3339),
	}
}

type BillingStatusResponse struct {
	Plan             PlanResponse `json:"plan"`
	CurrentPeriodEnd *string      `json:"current_period_end"`
	IsOverdue        bool         `json:"is_overdue"`
}

type InvoiceResponse struct {
	ID               string  `json:"id"`
	ShopID           string  `json:"shop_id"`
	Amount           string  `json:"amount"`
	Status           string  `json:"status"`
	BillingReason    *string `json:"billing_reason,omitempty"`
	HostedInvoiceUrl *string `json:"hosted_invoice_url,omitempty"`
	CreatedAt        string  `json:"created_at"`
}

func invoiceToResponse(inv db.ShopInvoice) InvoiceResponse {
	amount := ""
	if inv.Amount.Valid {
		amount = inv.Amount.Int.String()
		if inv.Amount.Exp != 0 {
			amount = fmt.Sprintf("%se%d", amount, inv.Amount.Exp)
		}
	}

	var reason *string
	if inv.BillingReason.Valid {
		r := inv.BillingReason.String
		reason = &r
	}
	var url *string
	if inv.HostedInvoiceUrl.Valid {
		u := inv.HostedInvoiceUrl.String
		url = &u
	}

	shopID := ""
	if inv.ShopID.Valid {
		shopID = uuid.UUID(inv.ShopID.Bytes).String()
	}

	return InvoiceResponse{
		ID:               inv.ID.String(),
		ShopID:           shopID,
		Amount:           amount,
		Status:           string(inv.Status),
		BillingReason:    reason,
		HostedInvoiceUrl: url,
		CreatedAt:        inv.CreatedAt.Time.Format(time.RFC3339),
	}
}

func (s *Service) ListPlans(ctx context.Context) ([]PlanResponse, error) {
	plans, err := s.q.ListPlans(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PlanResponse, len(plans))
	for i, p := range plans {
		out[i] = planToResponse(p)
	}
	return out, nil
}

func (s *Service) GetPlan(ctx context.Context, planID string) (PlanResponse, error) {
	id, err := uuid.Parse(planID)
	if err != nil {
		return PlanResponse{}, ErrInvalidPlan
	}
	p, err := s.q.GetPlan(ctx, id)
	if err != nil {
		return PlanResponse{}, ErrPlanNotFound
	}
	return planToResponse(p), nil
}

func (s *Service) GetShopBillingStatus(ctx context.Context, shopID string) (BillingStatusResponse, error) {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return BillingStatusResponse{}, ErrShopNotFound
	}

	shop, err := s.q.GetShop(ctx, id)
	if err != nil {
		return BillingStatusResponse{}, ErrShopNotFound
	}

	plan, err := s.q.GetPlan(ctx, shop.PlanID)
	if err != nil {
		return BillingStatusResponse{}, ErrPlanNotFound
	}

	var periodEnd *string
	if shop.CurrentPeriodEnd.Valid {
		t := shop.CurrentPeriodEnd.Time.Format(time.RFC3339)
		periodEnd = &t
	}

	isOverdue := shop.IsOverdue.Valid && shop.IsOverdue.Bool

	return BillingStatusResponse{
		Plan:             planToResponse(plan),
		CurrentPeriodEnd: periodEnd,
		IsOverdue:        isOverdue,
	}, nil
}

func (s *Service) ChangePlan(ctx context.Context, shopID string, planID string) (BillingStatusResponse, error) {
	sID, err := uuid.Parse(shopID)
	if err != nil {
		return BillingStatusResponse{}, ErrShopNotFound
	}
	pID, err := uuid.Parse(planID)
	if err != nil {
		return BillingStatusResponse{}, ErrInvalidPlan
	}

	_, err = s.q.GetPlan(ctx, pID)
	if err != nil {
		return BillingStatusResponse{}, ErrPlanNotFound
	}

	newPeriodEnd := time.Now().UTC().Add(30 * 24 * time.Hour)

	_, err = s.q.UpdateShop(ctx, db.UpdateShopParams{
		ID:     sID,
		PlanID: pgtype.UUID{Bytes: pID, Valid: true},
		CurrentPeriodEnd: pgtype.Timestamptz{
			Time:  newPeriodEnd,
			Valid: true,
		},
		IsOverdue: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		return BillingStatusResponse{}, err
	}

	return s.GetShopBillingStatus(ctx, shopID)
}

func (s *Service) MarkOverdue(ctx context.Context, shopID string) error {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}
	return s.q.UpdateShopBilling(ctx, db.UpdateShopBillingParams{
		ID:               id,
		CurrentPeriodEnd: pgtype.Timestamptz{}, // SQL COALESCE preserves existing value when Valid is false
		IsOverdue:        pgtype.Bool{Bool: true, Valid: true},
	})
}

func (s *Service) ClearOverdue(ctx context.Context, shopID string) error {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}
	return s.q.UpdateShopBilling(ctx, db.UpdateShopBillingParams{
		ID:               id,
		CurrentPeriodEnd: pgtype.Timestamptz{},
		IsOverdue:        pgtype.Bool{Bool: false, Valid: true},
	})
}

func (s *Service) ListInvoices(ctx context.Context, shopID string, page, pageSize int) ([]InvoiceResponse, error) {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := int32((page - 1) * pageSize)

	invoices, err := s.q.ListShopInvoices(ctx, db.ListShopInvoicesParams{
		ShopID: pgtype.UUID{Bytes: id, Valid: true},
		Limit:  int32(pageSize),
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	out := make([]InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		out[i] = invoiceToResponse(inv)
	}
	return out, nil
}

func (s *Service) CheckProductLimit(ctx context.Context, shopID string) error {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}

	shop, err := s.q.GetShop(ctx, id)
	if err != nil {
		return ErrShopNotFound
	}

	plan, err := s.q.GetPlan(ctx, shop.PlanID)
	if err != nil {
		return ErrPlanNotFound
	}

	if !plan.MaxProducts.Valid {
		return nil
	}

	count, err := s.q.CountProducts(ctx, db.CountProductsParams{
		ShopID: pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return err
	}

	if count >= int64(plan.MaxProducts.Int32) {
		return ErrProductLimit
	}
	return nil
}

func (s *Service) CheckStaffLimit(ctx context.Context, shopID string) error {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}

	shop, err := s.q.GetShop(ctx, id)
	if err != nil {
		return ErrShopNotFound
	}

	plan, err := s.q.GetPlan(ctx, shop.PlanID)
	if err != nil {
		return ErrPlanNotFound
	}

	if !plan.MaxStaffAccounts.Valid {
		return nil
	}

	count, err := s.q.CountShopStaff(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return err
	}

	if count >= int64(plan.MaxStaffAccounts.Int32) {
		return ErrStaffLimit
	}
	return nil
}

func (s *Service) GetPlanFeatures(ctx context.Context, shopID string) (map[string]any, error) {
	cacheKey := fmt.Sprintf("shop_plan_features:%s", shopID)
	if s.rdb != nil {
		if cachedJSON, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil && cachedJSON != "" {
			features := make(map[string]any)
			if err := json.Unmarshal([]byte(cachedJSON), &features); err == nil {
				return features, nil
			}
		}
	}

	id, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrShopNotFound
	}

	shop, err := s.q.GetShop(ctx, id)
	if err != nil {
		return nil, ErrShopNotFound
	}

	plan, err := s.q.GetPlan(ctx, shop.PlanID)
	if err != nil {
		return nil, ErrPlanNotFound
	}

	features := make(map[string]any)
	if len(plan.Features) > 0 {
		if err := json.Unmarshal(plan.Features, &features); err != nil {
			return features, nil
		}
	}

	if s.rdb != nil && len(features) > 0 {
		if encodedFeatures, err := json.Marshal(features); err == nil {
			s.rdb.Set(ctx, cacheKey, encodedFeatures, 5*time.Minute)
		}
	}

	return features, nil
}

type ChecklistItem struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type OnboardingChecklist struct {
	AllComplete bool            `json:"all_complete"`
	Progress    int             `json:"progress"`
	Items       []ChecklistItem `json:"items"`
}

func (s *Service) GetOnboardingChecklist(ctx context.Context, shopID string) (OnboardingChecklist, error) {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return OnboardingChecklist{}, ErrShopNotFound
	}

	shop, err := s.q.GetShop(ctx, id)
	if err != nil {
		return OnboardingChecklist{}, ErrShopNotFound
	}

	shopPgUUID := pgtype.UUID{Bytes: id, Valid: true}

	productCount, _ := s.q.CountProducts(ctx, db.CountProductsParams{ShopID: shopPgUUID})
	hasProduct := productCount > 0

	paymentMethods, _ := s.q.ListEnabledPaymentMethods(ctx, shopPgUUID)
	hasPayment := len(paymentMethods) > 0

	orderCount, _ := s.q.CountOrders(ctx, id)
	hasOrder := orderCount > 0

	items := []ChecklistItem{
		{
			Key:         "logo_uploaded",
			Title:       "Upload your shop logo",
			Description: "Set a favicon/logo to brand your storefront.",
			Completed:   shop.FaviconUrl.Valid && shop.FaviconUrl.String != "",
		},
		{
			Key:         "seo_configured",
			Title:       "Configure SEO settings",
			Description: "Set a title and description to improve search visibility.",
			Completed:   shop.SeoTitle.Valid && shop.SeoTitle.String != "",
		},
		{
			Key:         "first_product",
			Title:       "Add your first product",
			Description: "Create at least one product in your catalog.",
			Completed:   hasProduct,
		},
		{
			Key:         "payment_configured",
			Title:       "Set up a payment method",
			Description: "Enable at least one payment gateway to accept orders.",
			Completed:   hasPayment,
		},
		{
			Key:         "custom_domain",
			Title:       "Connect a custom domain",
			Description: "Point your own domain to your storefront.",
			Completed:   shop.CustomDomain.Valid && shop.CustomDomain.String != "",
		},
		{
			Key:         "first_order",
			Title:       "Receive your first order",
			Description: "You're all set when your first customer places an order.",
			Completed:   hasOrder,
		},
	}

	completed := 0
	for _, item := range items {
		if item.Completed {
			completed++
		}
	}

	progress := 0
	if len(items) > 0 {
		progress = (completed * 100) / len(items)
	}

	return OnboardingChecklist{
		AllComplete: completed == len(items),
		Progress:    progress,
		Items:       items,
	}, nil
}
