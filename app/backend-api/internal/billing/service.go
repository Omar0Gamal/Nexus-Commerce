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
)

// Sentinel errors
var (
	ErrPlanNotFound = errors.New("plan not found")
	ErrShopNotFound = errors.New("shop not found")
	ErrProductLimit = errors.New("product limit reached for current plan")
	ErrStaffLimit   = errors.New("staff seat limit reached for current plan")
	ErrInvalidPlan  = errors.New("invalid plan ID")
)

// Service handles subscription billing, plan management and plan-limit enforcement.
type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service {
	return &Service{q: q}
}

// Q exposes the underlying db.Queries for handler-layer audit writes.
func (s *Service) Q() *db.Queries { return s.q }

// PlanResponse is the public representation of a plan.
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

// BillingStatusResponse is the shop-facing billing overview.
type BillingStatusResponse struct {
	Plan             PlanResponse `json:"plan"`
	CurrentPeriodEnd *string      `json:"current_period_end"`
	IsOverdue        bool         `json:"is_overdue"`
}

// InvoiceResponse is the public representation of a shop invoice.
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

// ListPlans returns all available subscription plans ordered by price.
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

// GetPlan returns a single plan by ID.
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

// GetShopBillingStatus returns the current plan + billing state for a shop.
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

// ChangePlan updates the shop's subscription plan and resets the billing period.
func (s *Service) ChangePlan(ctx context.Context, shopID string, planID string) (BillingStatusResponse, error) {
	sID, err := uuid.Parse(shopID)
	if err != nil {
		return BillingStatusResponse{}, ErrShopNotFound
	}
	pID, err := uuid.Parse(planID)
	if err != nil {
		return BillingStatusResponse{}, ErrInvalidPlan
	}

	// Verify the plan exists.
	_, err = s.q.GetPlan(ctx, pID)
	if err != nil {
		return BillingStatusResponse{}, ErrPlanNotFound
	}

	// Set new period end to 30 days from now.
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

// MarkOverdue sets a shop's is_overdue flag to true.
func (s *Service) MarkOverdue(ctx context.Context, shopID string) error {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}
	return s.q.UpdateShopBilling(ctx, db.UpdateShopBillingParams{
		ID:               id,
		CurrentPeriodEnd: pgtype.Timestamptz{}, // keep unchanged — pass zero (SQL COALESCE will ignore)
		IsOverdue:        pgtype.Bool{Bool: true, Valid: true},
	})
}

// ClearOverdue clears a shop's is_overdue flag.
func (s *Service) ClearOverdue(ctx context.Context, shopID string) error {
	id, err := uuid.Parse(shopID)
	if err != nil {
		return ErrShopNotFound
	}
	return s.q.UpdateShopBilling(ctx, db.UpdateShopBillingParams{
		ID:               id,
		CurrentPeriodEnd: pgtype.Timestamptz{}, // keep unchanged
		IsOverdue:        pgtype.Bool{Bool: false, Valid: true},
	})
}

// ListInvoices returns the paginated invoice history for a shop.
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

// CheckProductLimit returns ErrProductLimit if the shop has reached its plan's
// max_products limit. A nil max_products means unlimited.
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

	// No limit set — unlimited.
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

// CheckStaffLimit returns ErrStaffLimit if the shop has reached its plan's
// max_staff_accounts limit. A nil max_staff_accounts means unlimited.
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

// GetPlanFeatures returns the decoded features map for a shop's current plan.
// Returns an empty map if features are not set.
func (s *Service) GetPlanFeatures(ctx context.Context, shopID string) (map[string]any, error) {
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
			return features, nil // return empty on parse failure
		}
	}
	return features, nil
}

// ChecklistItem represents a single onboarding step.
type ChecklistItem struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// OnboardingChecklist is the full checklist response.
type OnboardingChecklist struct {
	AllComplete bool            `json:"all_complete"`
	Progress    int             `json:"progress"` // 0–100 percentage
	Items       []ChecklistItem `json:"items"`
}

// GetOnboardingChecklist returns a completion checklist for initial shop setup.
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

	// Check first product
	productCount, _ := s.q.CountProducts(ctx, db.CountProductsParams{ShopID: shopPgUUID})
	hasProduct := productCount > 0

	// Check enabled payment method
	paymentMethods, _ := s.q.ListEnabledPaymentMethods(ctx, shopPgUUID)
	hasPayment := len(paymentMethods) > 0

	// Check first order
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
