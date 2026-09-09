ALTER TABLE shops
    DROP COLUMN IF EXISTS base_currency,
    DROP COLUMN IF EXISTS display_currencies,
    DROP COLUMN IF EXISTS default_locale,
    DROP COLUMN IF EXISTS timezone;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS currencies;
