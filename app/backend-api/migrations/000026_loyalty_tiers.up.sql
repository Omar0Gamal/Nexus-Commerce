-- ==================== Loyalty Tiers ====================

CREATE TABLE IF NOT EXISTS loyalty_tiers (
    id                   UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id              UUID          NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name                 TEXT          NOT NULL,
    min_lifetime_points  BIGINT        NOT NULL DEFAULT 0,
    benefits             JSONB         NOT NULL DEFAULT '{}',
    badge_label          TEXT          NOT NULL DEFAULT '',
    badge_color          TEXT          NOT NULL DEFAULT '#808080',
    position             INT           NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_loyalty_tiers_shop_pos ON loyalty_tiers(shop_id, position);
