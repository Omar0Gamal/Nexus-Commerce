ALTER TABLE coupons
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS min_order_amount,
    DROP COLUMN IF EXISTS usage_limit,
    DROP COLUMN IF EXISTS usage_count;
