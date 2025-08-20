package worker

import (
	"context"
	"time"

	"backend-api/internal/db"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgtype"
)

// LoyaltyExpiryWorker expires points for customers whose loyalty account
// hasn't been updated within the shop's configured expiry_months window.
type LoyaltyExpiryWorker struct {
	queries *db.Queries
	logger  *zap.Logger
}

func NewLoyaltyExpiryWorker(queries *db.Queries, logger *zap.Logger) *LoyaltyExpiryWorker {
	return &LoyaltyExpiryWorker{queries: queries, logger: logger}
}

// Run starts the nightly expiry loop.
func (w *LoyaltyExpiryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Process on startup
	w.expire(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.expire(ctx)
		}
	}
}

func (w *LoyaltyExpiryWorker) expire(ctx context.Context) {
	accounts, err := w.queries.GetAccountsWithExpiredPoints(ctx)
	if err != nil {
		w.logger.Error("loyalty expiry: failed to get expired accounts", zap.Error(err))
		return
	}
	if len(accounts) == 0 {
		return
	}

	w.logger.Info("loyalty expiry: expiring points", zap.Int("count", len(accounts)))

	for _, acc := range accounts {
		_, err := w.queries.RecordLoyaltyTransaction(ctx, db.RecordLoyaltyTransactionParams{
			ShopID:      acc.ShopID,
			CustomerID:  acc.CustomerID,
			PointsDelta: -acc.PointsBalance,
			Reason:      "expiry",
			ReferenceID: pgtype.UUID{},
		})
		if err != nil {
			w.logger.Error("loyalty expiry: failed to record expiry transaction",
				zap.String("customer_id", acc.CustomerID.String()),
				zap.Error(err),
			)
			continue
		}
		// Adjust the account balance to zero
		_, err = w.queries.UpsertLoyaltyAccount(ctx, db.UpsertLoyaltyAccountParams{
			ShopID:        acc.ShopID,
			CustomerID:    acc.CustomerID,
			PointsBalance: -acc.PointsBalance,
		})
		if err != nil {
			w.logger.Error("loyalty expiry: failed to update account balance",
				zap.String("customer_id", acc.CustomerID.String()),
				zap.Error(err),
			)
		}
	}
}
