-- ==================== Referral Config ====================

-- name: GetReferralConfig :one
SELECT * FROM referral_config WHERE shop_id = $1;

-- name: UpsertReferralConfig :one
INSERT INTO referral_config (
    shop_id, is_enabled, referrer_reward_type, referrer_reward_value,
    referee_reward_type, referee_reward_value, min_order_amount,
    max_referrals_per_customer, cookie_window_days, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()
)
ON CONFLICT (shop_id) DO UPDATE SET
    is_enabled                 = EXCLUDED.is_enabled,
    referrer_reward_type       = EXCLUDED.referrer_reward_type,
    referrer_reward_value      = EXCLUDED.referrer_reward_value,
    referee_reward_type        = EXCLUDED.referee_reward_type,
    referee_reward_value       = EXCLUDED.referee_reward_value,
    min_order_amount           = EXCLUDED.min_order_amount,
    max_referrals_per_customer = EXCLUDED.max_referrals_per_customer,
    cookie_window_days         = EXCLUDED.cookie_window_days,
    updated_at                 = NOW()
RETURNING *;
