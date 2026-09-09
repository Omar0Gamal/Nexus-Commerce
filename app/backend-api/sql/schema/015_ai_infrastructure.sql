-- Phase 11: AI Infrastructure

-- pg_trgm extension for Phase 10 trigram search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Platform-level AI configuration
CREATE TABLE ai_config (
    id                          SERIAL      PRIMARY KEY,
    primary_provider            TEXT        NOT NULL DEFAULT 'gemini',
    primary_api_key_encrypted   TEXT,
    primary_model               TEXT        NOT NULL DEFAULT 'gemini-2.0-flash-exp',
    fallback_provider           TEXT,
    fallback_api_key_encrypted  TEXT,
    fallback_model              TEXT,
    engine_base_url             TEXT,
    engine_api_key_encrypted    TEXT,
    max_tokens_per_request      INT         NOT NULL DEFAULT 2048,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Per-shop, per-feature AI call log
CREATE TABLE ai_usage_log (
    id               BIGSERIAL   PRIMARY KEY,
    shop_id          UUID        NOT NULL REFERENCES shops(id),
    feature          TEXT        NOT NULL,
    provider         TEXT        NOT NULL,
    model            TEXT        NOT NULL,
    input_tokens     INT         NOT NULL DEFAULT 0,
    output_tokens    INT         NOT NULL DEFAULT 0,
    latency_ms       INT         NOT NULL DEFAULT 0,
    cost_usd_micros  INT         NOT NULL DEFAULT 0,
    succeeded        BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_usage_log_shop_feature ON ai_usage_log(shop_id, feature, created_at DESC);
