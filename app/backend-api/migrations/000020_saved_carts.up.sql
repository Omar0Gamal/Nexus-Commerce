-- Phase 14: Abandoned Cart Recovery
CREATE TABLE saved_carts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    items           JSONB NOT NULL DEFAULT '[]',
    coupon_code     TEXT,
    total_cents     INT NOT NULL DEFAULT 0,
    recovery_token  TEXT,
    email_1_sent_at TIMESTAMPTZ,
    email_2_sent_at TIMESTAMPTZ,
    email_3_sent_at TIMESTAMPTZ,
    recovered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, customer_id)
);

CREATE INDEX idx_saved_carts_shop_updated ON saved_carts(shop_id, updated_at);
CREATE INDEX idx_saved_carts_abandoned ON saved_carts(shop_id, updated_at, recovered_at)
    WHERE recovered_at IS NULL;
