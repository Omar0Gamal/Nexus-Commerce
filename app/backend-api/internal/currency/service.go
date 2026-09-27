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

type Service struct {
	queries *db.Queries
	rdb     *redis.Client
}

func NewService(queries *db.Queries, rdb *redis.Client) *Service {
	return &Service{queries: queries, rdb: rdb}
}

func rateKey(from, to string) string {
	return fmt.Sprintf("rate:%s:%s", from, to)
}

func (s *Service) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	if s.rdb != nil {
		cachedRateStr, err := s.rdb.Get(ctx, rateKey(from, to)).Result()
		if err == nil {
			rate, parseErr := strconv.ParseFloat(cachedRateStr, 64)
			if parseErr == nil {
				return rate, nil
			}
		}
	}

	pgRate, err := s.queries.GetExchangeRate(ctx, db.GetExchangeRateParams{
		FromCurrency: from,
		ToCurrency:   to,
	})
	if err != nil {
		return 0, fmt.Errorf("currency.GetRate %s→%s: %w", from, to, err)
	}

	rateFloat, err := numericToFloat64(pgRate)
	if err != nil {
		return 0, err
	}

	// Cache TTL exceeds the 24-hour daily exchange rate sync cycle.
	if s.rdb != nil {
		_ = s.rdb.Set(ctx, rateKey(from, to), rateFloat, 25*60*60*1e9).Err()
	}
	return rateFloat, nil
}

func (s *Service) Convert(ctx context.Context, amountCents int64, from, to string) (int64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(float64(amountCents) * rate)), nil
}

func (s *Service) GetShopSettings(ctx context.Context, shopID uuid.UUID) (db.GetShopCurrencySettingsRow, error) {
	return s.queries.GetShopCurrencySettings(ctx, shopID)
}

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

func numericToFloat64(n pgtype.Numeric) (float64, error) {
	if !n.Valid {
		return 0, fmt.Errorf("numeric value is null")
	}
	parsedFloat, err := strconv.ParseFloat(n.Int.String(), 64)
	if err != nil {
		return 0, err
	}
	if n.Exp != 0 {
		parsedFloat *= math.Pow10(int(n.Exp))
	}
	return parsedFloat, nil
}
