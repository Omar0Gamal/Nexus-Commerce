-- Phase 16: Multi-Currency & Localization

-- Supported currencies reference table
CREATE TABLE currencies (
    code           TEXT PRIMARY KEY,   -- ISO 4217: "EGP", "USD", "EUR"
    name           TEXT NOT NULL,
    symbol         TEXT NOT NULL,
    decimal_digits SMALLINT NOT NULL DEFAULT 2,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE
);

INSERT INTO currencies VALUES
    ('EGP', 'Egyptian Pound',  'ج.م', 2, TRUE),
    ('USD', 'US Dollar',       '$',   2, TRUE),
    ('EUR', 'Euro',             '€',   2, TRUE),
    ('SAR', 'Saudi Riyal',     'ر.س', 2, TRUE),
    ('AED', 'UAE Dirham',      'د.إ', 2, TRUE),
    ('GBP', 'British Pound',   '£',   2, TRUE);

-- Exchange rates (refreshed daily by worker)
CREATE TABLE exchange_rates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency TEXT NOT NULL REFERENCES currencies(code),
    to_currency   TEXT NOT NULL REFERENCES currencies(code),
    rate          NUMERIC(18,8) NOT NULL,
    fetched_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (from_currency, to_currency)
);
CREATE INDEX idx_exchange_rates ON exchange_rates(from_currency, to_currency);

-- Shop locale settings
ALTER TABLE shops
    ADD COLUMN IF NOT EXISTS base_currency       TEXT NOT NULL DEFAULT 'EGP' REFERENCES currencies(code),
    ADD COLUMN IF NOT EXISTS display_currencies  TEXT[] NOT NULL DEFAULT ARRAY['EGP'],
    ADD COLUMN IF NOT EXISTS default_locale      TEXT NOT NULL DEFAULT 'ar-EG',
    ADD COLUMN IF NOT EXISTS timezone            TEXT NOT NULL DEFAULT 'Africa/Cairo';
