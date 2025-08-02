package currency

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"backend-api/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

// Service handles currency conversion and locale settings.
type Service struct {
	queries *db.Queries
	rdb     *redis.Client
}

func NewService(queries *db.Queries, rdb *redis.Client) *Service {
	return &Service{queries: queries, rdb: rdb}
}

// rateKey returns the Redis key for a currency pair.
func rateKey(from, to string) string {
	return fmt.Sprintf("rate:%s:%s", from, to)
}

// GetRate returns the exchange rate from → to.
// Checks Redis cache first, falls back to DB.
func (s *Service) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	// Try Redis cache.
	if s.rdb != nil {
		val, err := s.rdb.Get(ctx, rateKey(from, to)).Result()
		if err == nil {
			rate, parseErr := strconv.ParseFloat(val, 64)
			if parseErr == nil {
				return rate, nil
			}
		}
	}

	// Fall back to DB.
	pgRate, err := s.queries.GetExchangeRate(ctx, db.GetExchangeRateParams{
		FromCurrency: from,
		ToCurrency:   to,
	})
	if err != nil {
		return 0, fmt.Errorf("currency.GetRate %s→%s: %w", from, to, err)
	}

	// pgtype.Numeric → float64
	f, err := numericToFloat64(pgRate)
	if err != nil {
		return 0, err
	}

	// Cache for 25 hours.
	if s.rdb != nil {
		_ = s.rdb.Set(ctx, rateKey(from, to), f, 25*60*60*1e9).Err()
	}
	return f, nil
}

// Convert converts amountCents from one currency to another.
func (s *Service) Convert(ctx context.Context, amountCents int64, from, to string) (int64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(float64(amountCents) * rate)), nil
}

// GetShopSettings returns the locale/currency settings for a shop.
func (s *Service) GetShopSettings(ctx context.Context, shopID uuid.UUID) (db.GetShopCurrencySettingsRow, error) {
	return s.queries.GetShopCurrencySettings(ctx, shopID)
}

// UpdateShopLocale updates a shop's locale/currency configuration.
func (s *Service) UpdateShopLocale(ctx context.Context, shopID uuid.UUID, baseCurrency string, displayCurrencies []string, locale string, timezone string) (db.UpdateShopLocaleRow, error) {
	tz := pgtype.Text{}
	if timezone != "" {
		tz = pgtype.Text{String: timezone, Valid: true}
	}
	return s.queries.UpdateShopLocale(ctx, db.UpdateShopLocaleParams{
		ID:                shopID,
		BaseCurrency:      baseCurrency,
		DisplayCurrencies: displayCurrencies,
		DefaultLocale:     locale,
		Timezone:          tz,
	})
}

// numericToFloat64 converts pgtype.Numeric to float64.
func numericToFloat64(n pgtype.Numeric) (float64, error) {
	if !n.Valid {
		return 0, fmt.Errorf("numeric value is null")
	}
	f, err := strconv.ParseFloat(n.Int.String(), 64)
	if err != nil {
		return 0, err
	}
	if n.Exp != 0 {
		f *= math.Pow10(int(n.Exp))
	}
	return f, nil
}
