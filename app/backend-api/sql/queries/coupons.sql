-- ==================== Coupons ====================

-- name: CreateCoupon :one
INSERT INTO coupons (
    shop_id, code, type, value, is_active, expires_at, min_order_amount, usage_limit
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetCoupon :one
SELECT * FROM coupons
WHERE id = $1 AND shop_id = $2;

-- name: GetCouponByCode :one
SELECT * FROM coupons
WHERE code = $1 AND shop_id = $2;

-- name: ListCoupons :many
SELECT * FROM coupons
WHERE shop_id = $1
ORDER BY code ASC;

-- name: ListActiveCoupons :many
SELECT * FROM coupons
WHERE shop_id = $1 AND is_active = TRUE
ORDER BY code ASC;

-- name: UpdateCoupon :one
UPDATE coupons SET
    code = COALESCE(sqlc.narg('code'), code),
    type = COALESCE(sqlc.narg('type'), type),
    value = COALESCE(sqlc.narg('value'), value),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    min_order_amount = COALESCE(sqlc.narg('min_order_amount'), min_order_amount),
    usage_limit = COALESCE(sqlc.narg('usage_limit'), usage_limit)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: IncrementCouponUsage :one
UPDATE coupons
SET usage_count = usage_count + 1
WHERE id = $1 AND shop_id = $2
RETURNING usage_count;

-- name: ToggleCoupon :one
UPDATE coupons SET is_active = NOT is_active
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: DeleteCoupon :exec
DELETE FROM coupons WHERE id = $1 AND shop_id = $2;

-- name: ValidateCoupon :one
-- Returns the coupon only when active, not expired, and under usage limit.
SELECT * FROM coupons
WHERE code = $1
  AND shop_id = $2
  AND is_active = TRUE
  AND (expires_at IS NULL OR expires_at > NOW())
  AND (usage_limit IS NULL OR usage_count < usage_limit);
