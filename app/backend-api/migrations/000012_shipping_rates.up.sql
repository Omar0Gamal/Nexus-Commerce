-- Add shipping_rates table and shipping_fee to orders

CREATE TYPE shipping_rate_type AS ENUM ('flat', 'weight_based');

CREATE TABLE shipping_rates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id     UUID NOT NULL REFERENCES shipping_zones(id) ON DELETE CASCADE,
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    rate_type   shipping_rate_type NOT NULL DEFAULT 'flat',
    -- flat rate: base_rate is the fee
    -- weight_based: base_rate + (rate_per_kg * weight_kg)
    base_rate   DECIMAL(10, 2) NOT NULL DEFAULT 0,
    rate_per_kg DECIMAL(10, 2),
    min_order_free_shipping DECIMAL(10, 2),  -- free shipping above this order total
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_shipping_rates_zone ON shipping_rates(zone_id);
CREATE INDEX idx_shipping_rates_shop ON shipping_rates(shop_id);

-- Add shipping_fee column to orders
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS shipping_fee DECIMAL(10, 2) NOT NULL DEFAULT 0;
