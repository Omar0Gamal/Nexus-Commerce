CREATE TABLE ab_tests (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id       UUID        NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name          TEXT        NOT NULL,
    description   TEXT,
    variant_a     JSONB       NOT NULL DEFAULT '{}',
    variant_b     JSONB       NOT NULL DEFAULT '{}',
    metric        TEXT        NOT NULL CHECK (metric IN ('click','cart_add','purchase')),
    traffic_split INT         NOT NULL DEFAULT 50 CHECK (traffic_split BETWEEN 1 AND 99),
    status        TEXT        NOT NULL DEFAULT 'active' CHECK (status IN ('active','paused','concluded')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    concluded_at  TIMESTAMPTZ
);

CREATE INDEX idx_ab_tests_shop_status ON ab_tests(shop_id, status);
