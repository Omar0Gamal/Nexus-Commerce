-- Phase 16: Multi-Currency & Localization
-- mirrors migrations/000022_currency_localization.up.sql for sqlc

CREATE TABLE currencies (
    code           TEXT PRIMARY KEY,
    name           TEXT NOT NULL,
    symbol         TEXT NOT NULL,
    decimal_digits SMALLINT NOT NULL DEFAULT 2,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE exchange_rates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency TEXT NOT NULL REFERENCES currencies(code),
    to_currency   TEXT NOT NULL REFERENCES currencies(code),
    rate          NUMERIC(18,8) NOT NULL,
    fetched_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (from_currency, to_currency)
);

ALTER TABLE shops
    ADD COLUMN IF NOT EXISTS base_currency       TEXT NOT NULL DEFAULT 'EGP',
    ADD COLUMN IF NOT EXISTS display_currencies  TEXT[] NOT NULL DEFAULT ARRAY['EGP'],
    ADD COLUMN IF NOT EXISTS default_locale      TEXT NOT NULL DEFAULT 'ar-EG';
