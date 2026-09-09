-- ==================== Referral Config ====================

CREATE TABLE IF NOT EXISTS referral_config (
    shop_id                  UUID          PRIMARY KEY REFERENCES shops(id) ON DELETE CASCADE,
    is_enabled               BOOLEAN       NOT NULL DEFAULT false,
    referrer_reward_type     TEXT          NOT NULL DEFAULT 'points',
    referrer_reward_value    NUMERIC(10,2) NOT NULL DEFAULT 100,
    referee_reward_type      TEXT          NOT NULL DEFAULT 'discount',
    referee_reward_value     NUMERIC(10,2) NOT NULL DEFAULT 50,
    min_order_amount         NUMERIC(10,2),
    max_referrals_per_customer INT,
    cookie_window_days       INT           NOT NULL DEFAULT 30,
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
