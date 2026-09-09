CREATE TABLE loyalty_tiers (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id              UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    min_lifetime_points  BIGINT NOT NULL DEFAULT 0,
    benefits             JSONB,
    badge_label          TEXT,
    badge_color          TEXT,
    position             INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_loyalty_tiers_shop_pos ON loyalty_tiers(shop_id, position);
