package subscriptions

import (
	"context"
	"errors"
	"time"

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
	ErrConflict    = errors.New("customer already has an active subscription to this plan")
)

// Service handles subscription business logic.
type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreatePlan(ctx context.Context, shopID string, req CreatePlanRequest) (*PlanResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	productID := pgtype.UUID{}
	if req.ProductID != "" {
		if id, err := uuid.Parse(req.ProductID); err == nil {
			productID = pgtype.UUID{Bytes: id, Valid: true}
		}
	}

	plan, err := s.queries.CreateSubscriptionPlan(ctx, db.CreateSubscriptionPlanParams{
		ShopID:       shopUUID,
		Name:         req.Name,
		Description:  pgtype.Text{String: req.Description, Valid: req.Description != ""},
		ProductID:    productID,
		Price:        pgutil.Float64ToNumeric(req.Price),
		BillingCycle: req.BillingCycle,
		TrialDays:    req.TrialDays,
		IsActive:     isActive,
	})
	if err != nil {
		return nil, err
	}
	return mapPlan(plan), nil
}

func (s *Service) GetPlan(ctx context.Context, shopID, planID string) (*PlanResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	plan, err := s.queries.GetSubscriptionPlan(ctx, db.GetSubscriptionPlanParams{
		ID:     planUUID,
		ShopID: shopUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return mapPlan(plan), nil
}

func (s *Service) ListPlans(ctx context.Context, shopID string) ([]PlanResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	plans, err := s.queries.ListSubscriptionPlans(ctx, shopUUID)
	if err != nil {
		return nil, err
	}
	result := make([]PlanResponse, len(plans))
	for i, p := range plans {
		result[i] = *mapPlan(p)
	}
	return result, nil
}

func (s *Service) UpdatePlan(ctx context.Context, shopID, planID string, req UpdatePlanRequest) (*PlanResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	existing, err := s.queries.GetSubscriptionPlan(ctx, db.GetSubscriptionPlanParams{ID: planUUID, ShopID: shopUUID})
	if err != nil {
		return nil, mapErr(err)
	}

	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}
	desc := existing.Description
	if req.Description != "" {
		desc = pgtype.Text{String: req.Description, Valid: true}
	}
	price := existing.Price
	if req.Price > 0 {
		price = pgutil.Float64ToNumeric(req.Price)
	}
	cycle := existing.BillingCycle
	if req.BillingCycle != "" {
		cycle = req.BillingCycle
	}
	trialDays := existing.TrialDays
	if req.TrialDays >= 0 {
		trialDays = req.TrialDays
	}
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	plan, err := s.queries.UpdateSubscriptionPlan(ctx, db.UpdateSubscriptionPlanParams{
		ID:           planUUID,
		ShopID:       shopUUID,
		Name:         name,
		Description:  desc,
		Price:        price,
		BillingCycle: cycle,
		TrialDays:    trialDays,
		IsActive:     isActive,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return mapPlan(plan), nil
}

func (s *Service) DeletePlan(ctx context.Context, shopID, planID string) error {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return ErrInvalidUUID
	}
	planUUID, err := uuid.Parse(planID)
	if err != nil {
		return ErrInvalidUUID
	}
	return s.queries.DeleteSubscriptionPlan(ctx, db.DeleteSubscriptionPlanParams{
		ID:     planUUID,
		ShopID: shopUUID,
	})
}

func (s *Service) Subscribe(ctx context.Context, shopID, customerID string, req SubscribeRequest) (*SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	planUUID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	plan, err := s.queries.GetSubscriptionPlan(ctx, db.GetSubscriptionPlanParams{
		ID:     planUUID,
		ShopID: shopUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}

	now := time.Now().UTC()
	status := "active"
	trialEnd := pgtype.Timestamptz{}
	nextBilling := now

	if plan.TrialDays > 0 {
		status = "trialing"
		te := now.AddDate(0, 0, int(plan.TrialDays))
		trialEnd = pgtype.Timestamptz{Time: te, Valid: true}
		nextBilling = te
	}

	periodEnd := nextBillingDate(now, plan.BillingCycle)
	next := nextBillingDate(nextBilling, plan.BillingCycle)

	sub, err := s.queries.CreateSubscription(ctx, db.CreateSubscriptionParams{
		ShopID:             shopUUID,
		CustomerID:         customerUUID,
		PlanID:             planUUID,
		Status:             status,
		CurrentPeriodStart: pgtype.Timestamptz{Time: now, Valid: true},
		CurrentPeriodEnd:   pgtype.Timestamptz{Time: periodEnd, Valid: true},
		NextBillingAt:      pgtype.Timestamptz{Time: next, Valid: true},
		TrialEnd:           trialEnd,
	})
	if err != nil {
		return nil, err
	}
	return mapSubscription(sub, plan.Name, pgutil.NumericToString(plan.Price), plan.BillingCycle), nil
}

func (s *Service) GetSubscription(ctx context.Context, shopID, subID string) (*SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	subUUID, err := uuid.Parse(subID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	row, err := s.queries.GetSubscription(ctx, db.GetSubscriptionParams{
		ID:     subUUID,
		ShopID: shopUUID,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	sub := db.Subscription{
		ID:                 row.ID,
		ShopID:             row.ShopID,
		CustomerID:         row.CustomerID,
		PlanID:             row.PlanID,
		Status:             row.Status,
		CurrentPeriodStart: row.CurrentPeriodStart,
		CurrentPeriodEnd:   row.CurrentPeriodEnd,
		NextBillingAt:      row.NextBillingAt,
		TrialEnd:           row.TrialEnd,
		CancelledAt:        row.CancelledAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
	return mapSubscription(sub, row.PlanName, pgutil.NumericToString(row.PlanPrice), row.BillingCycle), nil
}

func (s *Service) ListCustomerSubscriptions(ctx context.Context, shopID, customerID string) ([]SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	rows, err := s.queries.ListCustomerSubscriptions(ctx, db.ListCustomerSubscriptionsParams{
		CustomerID: customerUUID,
		ShopID:     shopUUID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]SubscriptionResponse, len(rows))
	for i, r := range rows {
		sub := db.Subscription{
			ID: r.ID, ShopID: r.ShopID, CustomerID: r.CustomerID, PlanID: r.PlanID,
			Status: r.Status, CurrentPeriodStart: r.CurrentPeriodStart,
			CurrentPeriodEnd: r.CurrentPeriodEnd, NextBillingAt: r.NextBillingAt,
			TrialEnd: r.TrialEnd, CancelledAt: r.CancelledAt, CreatedAt: r.CreatedAt,
		}
		result[i] = *mapSubscription(sub, r.PlanName, pgutil.NumericToString(r.PlanPrice), r.BillingCycle)
	}
	return result, nil
}

func (s *Service) ListShopSubscriptions(ctx context.Context, shopID string, page, perPage int) ([]SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	rows, err := s.queries.ListShopSubscriptions(ctx, db.ListShopSubscriptionsParams{
		ShopID: shopUUID,
		Lim:    int32(perPage),
		Off:    int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, err
	}
	result := make([]SubscriptionResponse, len(rows))
	for i, r := range rows {
		sub := db.Subscription{
			ID: r.ID, ShopID: r.ShopID, CustomerID: r.CustomerID, PlanID: r.PlanID,
			Status: r.Status, CurrentPeriodStart: r.CurrentPeriodStart,
			CurrentPeriodEnd: r.CurrentPeriodEnd, NextBillingAt: r.NextBillingAt,
			TrialEnd: r.TrialEnd, CancelledAt: r.CancelledAt, CreatedAt: r.CreatedAt,
		}
		result[i] = *mapSubscription(sub, r.PlanName, pgutil.NumericToString(r.PlanPrice), r.BillingCycle)
	}
	return result, nil
}

func (s *Service) CancelSubscription(ctx context.Context, shopID, subID string) (*SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	subUUID, err := uuid.Parse(subID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	now := time.Now().UTC()
	sub, err := s.queries.UpdateSubscriptionStatus(ctx, db.UpdateSubscriptionStatusParams{
		ID:          subUUID,
		ShopID:      shopUUID,
		Status:      "cancelled",
		CancelledAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return mapSubscription(sub, "", "", ""), nil
}

func (s *Service) PauseSubscription(ctx context.Context, shopID, subID string) (*SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	subUUID, err := uuid.Parse(subID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	sub, err := s.queries.UpdateSubscriptionStatus(ctx, db.UpdateSubscriptionStatusParams{
		ID:          subUUID,
		ShopID:      shopUUID,
		Status:      "paused",
		CancelledAt: pgtype.Timestamptz{},
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return mapSubscription(sub, "", "", ""), nil
}

func (s *Service) ResumeSubscription(ctx context.Context, shopID, subID string) (*SubscriptionResponse, error) {
	shopUUID, err := uuid.Parse(shopID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	subUUID, err := uuid.Parse(subID)
	if err != nil {
		return nil, ErrInvalidUUID
	}
	sub, err := s.queries.UpdateSubscriptionStatus(ctx, db.UpdateSubscriptionStatusParams{
		ID:          subUUID,
		ShopID:      shopUUID,
		Status:      "active",
		CancelledAt: pgtype.Timestamptz{},
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return mapSubscription(sub, "", "", ""), nil
}

func mapPlan(p db.SubscriptionPlan) *PlanResponse {
	desc := ""
	if p.Description.Valid {
		desc = p.Description.String
	}
	var productID *string
	if p.ProductID.Valid {
		s := uuid.UUID(p.ProductID.Bytes).String()
		productID = &s
	}
	return &PlanResponse{
		ID:           p.ID.String(),
		ShopID:       p.ShopID.String(),
		Name:         p.Name,
		Description:  desc,
		ProductID:    productID,
		Price:        pgutil.NumericToString(p.Price),
		BillingCycle: p.BillingCycle,
		TrialDays:    p.TrialDays,
		IsActive:     p.IsActive,
		CreatedAt:    p.CreatedAt.Time.Format(time.RFC3339),
	}
}

func mapSubscription(s db.Subscription, planName, planPrice, billingCycle string) *SubscriptionResponse {
	var trialEnd *string
	if s.TrialEnd.Valid {
		t := s.TrialEnd.Time.Format(time.RFC3339)
		trialEnd = &t
	}
	var cancelledAt *string
	if s.CancelledAt.Valid {
		t := s.CancelledAt.Time.Format(time.RFC3339)
		cancelledAt = &t
	}
	return &SubscriptionResponse{
		ID:                 s.ID.String(),
		ShopID:             s.ShopID.String(),
		CustomerID:         s.CustomerID.String(),
		PlanID:             s.PlanID.String(),
		PlanName:           planName,
		PlanPrice:          planPrice,
		BillingCycle:       billingCycle,
		Status:             s.Status,
		CurrentPeriodStart: s.CurrentPeriodStart.Time.Format(time.RFC3339),
		CurrentPeriodEnd:   s.CurrentPeriodEnd.Time.Format(time.RFC3339),
		NextBillingAt:      s.NextBillingAt.Time.Format(time.RFC3339),
		TrialEnd:           trialEnd,
		CancelledAt:        cancelledAt,
		CreatedAt:          s.CreatedAt.Time.Format(time.RFC3339),
	}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
