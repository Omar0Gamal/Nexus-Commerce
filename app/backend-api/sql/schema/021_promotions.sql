CREATE TABLE flash_sales (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id        UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    discount_type  TEXT NOT NULL CHECK (discount_type IN ('percentage','fixed')),
    discount_value NUMERIC(10,2) NOT NULL,
    starts_at      TIMESTAMPTZ NOT NULL,
    ends_at        TIMESTAMPTZ NOT NULL,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    max_uses       INT,
    uses_count     INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_flash_sales_shop_active ON flash_sales(shop_id, is_active, starts_at, ends_at);

CREATE TABLE flash_sale_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flash_sale_id UUID NOT NULL REFERENCES flash_sales(id) ON DELETE CASCADE,
    shop_id       UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id    UUID REFERENCES product_variants(id) ON DELETE CASCADE,
    sale_price_cents INT NOT NULL,
    UNIQUE (flash_sale_id, product_id, variant_id)
);
CREATE INDEX idx_flash_sale_items_product ON flash_sale_items(shop_id, product_id);

CREATE TABLE tiered_discounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    applies_to  TEXT NOT NULL CHECK (applies_to IN ('order','product','category')),
    target_id   UUID,
    tiers       JSONB NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tiered_discounts_shop ON tiered_discounts(shop_id, is_active);

CREATE TABLE loyalty_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    points_balance  INT NOT NULL DEFAULT 0 CHECK (points_balance >= 0),
    lifetime_points INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, customer_id)
);

CREATE TABLE loyalty_transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id      UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_id  UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    points_delta INT NOT NULL,
    reason       TEXT NOT NULL CHECK (reason IN ('purchase','redeem','referral','adjustment','expiry')),
    reference_id UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_loyalty_tx_customer ON loyalty_transactions(shop_id, customer_id, created_at DESC);

CREATE TABLE loyalty_programs (
    shop_id           UUID PRIMARY KEY REFERENCES shops(id) ON DELETE CASCADE,
    points_per_egp    NUMERIC(6,2) NOT NULL DEFAULT 1.0,
    egp_per_point     NUMERIC(6,4) NOT NULL DEFAULT 0.01,
    min_redeem_points INT NOT NULL DEFAULT 100,
    max_redeem_pct    SMALLINT NOT NULL DEFAULT 20,
    expiry_months     SMALLINT,
    is_active         BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE referral_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    code        TEXT NOT NULL UNIQUE,
    uses_count  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_referral_codes_shop ON referral_codes(shop_id, customer_id);

CREATE TABLE referral_conversions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id                 UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    referrer_id             UUID NOT NULL REFERENCES customers(id),
    referred_id             UUID NOT NULL REFERENCES customers(id),
    order_id                UUID REFERENCES orders(id),
    referrer_reward_pts     INT NOT NULL DEFAULT 0,
    referred_discount_cents INT NOT NULL DEFAULT 0,
    converted_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
