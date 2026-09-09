-- ==================== Marketing & Logistics ====================

CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    type discount_type NOT NULL,
    value DECIMAL(10, 2) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    min_order_amount DECIMAL(10, 2),
    usage_limit INTEGER,
    usage_count INTEGER NOT NULL DEFAULT 0,
    UNIQUE(shop_id, code)
);

CREATE TABLE shipping_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    regions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE shipping_rate_type AS ENUM ('flat', 'weight_based');

CREATE TABLE shipping_rates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id     UUID NOT NULL REFERENCES shipping_zones(id) ON DELETE CASCADE,
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    rate_type   shipping_rate_type NOT NULL DEFAULT 'flat',
    base_rate   DECIMAL(10, 2) NOT NULL DEFAULT 0,
    rate_per_kg DECIMAL(10, 2),
    min_order_free_shipping DECIMAL(10, 2),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_shipping_rates_zone ON shipping_rates(zone_id);
CREATE INDEX idx_shipping_rates_shop ON shipping_rates(shop_id);
