-- name: GetActiveCurrencies :many
SELECT * FROM currencies WHERE is_active = TRUE ORDER BY code;

-- name: GetExchangeRate :one
SELECT rate FROM exchange_rates
WHERE from_currency = $1 AND to_currency = $2;

-- name: UpsertExchangeRate :exec
INSERT INTO exchange_rates (from_currency, to_currency, rate, fetched_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (from_currency, to_currency) DO UPDATE
SET rate = EXCLUDED.rate, fetched_at = NOW();

-- name: GetShopCurrencySettings :one
SELECT base_currency, display_currencies, default_locale, timezone
FROM shops WHERE id = $1;

-- name: UpdateShopLocale :one
UPDATE shops
SET base_currency       = $2,
    display_currencies  = $3,
    default_locale      = $4,
    timezone            = $5
WHERE id = $1
RETURNING id, base_currency, display_currencies, default_locale, timezone;
