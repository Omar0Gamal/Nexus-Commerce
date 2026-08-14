-- Phase 14 schema for sqlc
CREATE TABLE saved_carts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL,
    customer_id     UUID NOT NULL,
    items           JSONB NOT NULL DEFAULT '[]',
    coupon_code     TEXT,
    total_cents     INT NOT NULL DEFAULT 0,
    recovery_token  TEXT,
    email_1_sent_at TIMESTAMPTZ,
    email_2_sent_at TIMESTAMPTZ,
    email_3_sent_at TIMESTAMPTZ,
    recovered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
