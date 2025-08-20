package worker

import (
	"context"
	"time"

	"backend-api/internal/db"
	"backend-api/internal/shared/pgutil"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgtype"
)

// SubscriptionRenewalWorker processes due subscriptions and creates renewal orders.
type SubscriptionRenewalWorker struct {
	queries *db.Queries
	logger  *zap.Logger
}

func NewSubscriptionRenewalWorker(queries *db.Queries, logger *zap.Logger) *SubscriptionRenewalWorker {
	return &SubscriptionRenewalWorker{queries: queries, logger: logger}
}

// Run starts the worker loop, processing due subscriptions every hour.
func (w *SubscriptionRenewalWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Process immediately on start
	w.process(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *SubscriptionRenewalWorker) process(ctx context.Context) {
	due, err := w.queries.GetDueSubscriptions(ctx)
	if err != nil {
		w.logger.Error("subscription renewal: failed to fetch due subscriptions", zap.Error(err))
		return
	}
	if len(due) == 0 {
		return
	}

	w.logger.Info("subscription renewal: processing due subscriptions", zap.Int("count", len(due)))

	for _, sub := range due {
		if err := w.renewSubscription(ctx, sub); err != nil {
			w.logger.Error("subscription renewal: failed to renew",
				zap.String("subscription_id", sub.ID.String()),
				zap.Error(err),
			)
		}
	}
}

func (w *SubscriptionRenewalWorker) renewSubscription(ctx context.Context, sub db.GetDueSubscriptionsRow) error {
	now := time.Now().UTC()

	// Calculate next billing period based on billing_cycle
	var nextBilling time.Time
	switch sub.BillingCycle {
	case "weekly":
		nextBilling = now.AddDate(0, 0, 7)
	case "quarterly":
		nextBilling = now.AddDate(0, 3, 0)
	case "annual":
		nextBilling = now.AddDate(1, 0, 0)
	default: // monthly
		nextBilling = now.AddDate(0, 1, 0)
	}

	// Record the subscription order (order_id null = pending fulfillment)
	_, err := w.queries.CreateSubscriptionOrder(ctx, db.CreateSubscriptionOrderParams{
		SubscriptionID: sub.ID,
		OrderID:        pgtype.UUID{},
		BillingDate:    pgtype.Timestamptz{Time: now, Valid: true},
		Status:         "pending",
	})
	if err != nil {
		return err
	}

	// Advance the billing window
	periodStart := pgtype.Timestamptz{Time: now, Valid: true}
	periodEnd := pgtype.Timestamptz{Time: nextBilling, Valid: true}
	nextBillingAt := pgtype.Timestamptz{Time: nextBilling, Valid: true}
	_, err = w.queries.UpdateSubscriptionBilling(ctx, db.UpdateSubscriptionBillingParams{
		ID:                 sub.ID,
		NextBillingAt:      nextBillingAt,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		Status:             "active",
	})
	if err != nil {
		return err
	}

	w.logger.Info("subscription renewal: renewed subscription",
		zap.String("subscription_id", sub.ID.String()),
		zap.String("customer_id", sub.CustomerID.String()),
		zap.String("price", pgutil.NumericToString(sub.PlanPrice)),
		zap.Time("next_billing_at", nextBilling),
	)
	return nil
}
