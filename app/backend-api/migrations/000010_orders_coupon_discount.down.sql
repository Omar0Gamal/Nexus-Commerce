ALTER TABLE orders
    DROP COLUMN IF EXISTS coupon_id,
    DROP COLUMN IF EXISTS discount_amount;
