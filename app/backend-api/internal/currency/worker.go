package currency

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"backend-api/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type RateWorker struct {
	svc    *Service
	logger *zap.Logger
}

func NewRateWorker(svc *Service, logger *zap.Logger) *RateWorker {
	return &RateWorker{svc: svc, logger: logger}
}

func (w *RateWorker) Run(ctx context.Context) {
	w.logger.Info("[currency] rate worker started")

	w.refresh(ctx)

	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 6, 0, 0, 0, time.UTC)
		sleep := time.Until(next)

		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
			w.refresh(ctx)
		}
	}
}

// Fallback rates when external rate provider is not configured.
func (w *RateWorker) refresh(ctx context.Context) {
	rates := hardcodedRates()

	for pair, rate := range rates {
		if err := w.upsert(ctx, pair[0], pair[1], rate); err != nil {
			w.logger.Error("upsert exchange rate",
				zap.String("from", pair[0]),
				zap.String("to", pair[1]),
				zap.Error(err))
		}
		if w.svc.rdb != nil {
			if err := w.svc.rdb.Set(ctx, rateKey(pair[0], pair[1]), fmt.Sprintf("%f", rate), 25*time.Hour).Err(); err != nil {
				w.logger.Warn("cache exchange rate",
					zap.String("from", pair[0]),
					zap.String("to", pair[1]),
					zap.Error(err),
				)
			}
		}
	}
	w.logger.Info("[currency] exchange rates refreshed", zap.Int("pairs", len(rates)))
}

func (w *RateWorker) upsert(ctx context.Context, from, to string, rate float64) error {
	num := pgtype.Numeric{Int: big.NewInt(int64(rate * 1e8)), Exp: -8, Valid: true}
	return w.svc.queries.UpsertExchangeRate(ctx, db.UpsertExchangeRateParams{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         num,
	})
}

func hardcodedRates() map[[2]string]float64 {
	egpToUSD := 0.0206
	egpToEUR := 0.0188
	egpToSAR := 0.0772
	egpToAED := 0.0756
	egpToGBP := 0.0163

	return map[[2]string]float64{
		{"EGP", "USD"}: egpToUSD,
		{"EGP", "EUR"}: egpToEUR,
		{"EGP", "SAR"}: egpToSAR,
		{"EGP", "AED"}: egpToAED,
		{"EGP", "GBP"}: egpToGBP,
		{"USD", "EGP"}: 1.0 / egpToUSD,
		{"EUR", "EGP"}: 1.0 / egpToEUR,
		{"SAR", "EGP"}: 1.0 / egpToSAR,
		{"AED", "EGP"}: 1.0 / egpToAED,
		{"GBP", "EGP"}: 1.0 / egpToGBP,
	}
}
