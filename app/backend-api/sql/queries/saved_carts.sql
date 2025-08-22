-- ==================== Saved / Abandoned Carts ====================

-- name: UpsertSavedCart :one
INSERT INTO saved_carts (shop_id, customer_id, items, coupon_code, total_cents, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (shop_id, customer_id) DO UPDATE
SET items       = EXCLUDED.items,
    coupon_code = EXCLUDED.coupon_code,
    total_cents = EXCLUDED.total_cents,
    updated_at  = NOW()
RETURNING id, shop_id, customer_id, items, coupon_code, total_cents, recovery_token,
          email_1_sent_at, email_2_sent_at, email_3_sent_at, recovered_at, created_at, updated_at;

-- name: GetSavedCart :one
SELECT id, shop_id, customer_id, items, coupon_code, total_cents, recovery_token,
       email_1_sent_at, email_2_sent_at, email_3_sent_at, recovered_at, created_at, updated_at
FROM saved_carts
WHERE shop_id = $1 AND customer_id = $2;

-- name: DeleteSavedCart :exec
DELETE FROM saved_carts WHERE shop_id = $1 AND customer_id = $2;

-- name: SetRecoveryToken :exec
UPDATE saved_carts SET recovery_token = $3
WHERE shop_id = $1 AND customer_id = $2;

-- name: GetCartByRecoveryToken :one
SELECT sc.id, sc.shop_id, sc.customer_id, sc.items, sc.coupon_code, sc.total_cents,
       sc.recovery_token, sc.email_1_sent_at, sc.email_2_sent_at, sc.email_3_sent_at,
       sc.recovered_at, sc.created_at, sc.updated_at,
       c.email AS customer_email
FROM saved_carts sc
JOIN customers c ON c.id = sc.customer_id
WHERE sc.recovery_token = $1
  AND sc.recovered_at IS NULL
LIMIT 1;

-- name: MarkCartRecovered :exec
UPDATE saved_carts
SET recovered_at = NOW(), recovery_token = NULL
WHERE shop_id = $1 AND customer_id = $2;

-- name: MarkCartEmail1Sent :exec
UPDATE saved_carts SET email_1_sent_at = NOW()
WHERE shop_id = $1 AND id = $2;

-- name: MarkCartEmail2Sent :exec
UPDATE saved_carts SET email_2_sent_at = NOW()
WHERE shop_id = $1 AND id = $2;

-- name: MarkCartEmail3Sent :exec
UPDATE saved_carts SET email_3_sent_at = NOW()
WHERE shop_id = $1 AND id = $2;

-- name: GetCartsForEmail1 :many
-- Carts idle for 1-2 hours, email_1 not yet sent, not recovered.
SELECT sc.id, sc.shop_id, sc.customer_id, sc.items, sc.coupon_code, sc.total_cents,
       sc.recovery_token, sc.email_1_sent_at, sc.email_2_sent_at, sc.email_3_sent_at,
       sc.recovered_at, sc.created_at, sc.updated_at,
       c.email AS customer_email, c.first_name AS customer_first_name
FROM saved_carts sc
JOIN customers c ON c.id = sc.customer_id
WHERE sc.updated_at BETWEEN NOW() - INTERVAL '2 hours' AND NOW() - INTERVAL '1 hour'
  AND sc.email_1_sent_at IS NULL
  AND sc.recovered_at IS NULL
  AND jsonb_array_length(sc.items) > 0;

-- name: GetCartsForEmail2 :many
-- Carts idle 23-25 hours after email_1 was sent, email_2 not yet sent.
SELECT sc.id, sc.shop_id, sc.customer_id, sc.items, sc.coupon_code, sc.total_cents,
       sc.recovery_token, sc.email_1_sent_at, sc.email_2_sent_at, sc.email_3_sent_at,
       sc.recovered_at, sc.created_at, sc.updated_at,
       c.email AS customer_email, c.first_name AS customer_first_name
FROM saved_carts sc
JOIN customers c ON c.id = sc.customer_id
WHERE sc.email_1_sent_at BETWEEN NOW() - INTERVAL '25 hours' AND NOW() - INTERVAL '23 hours'
  AND sc.email_2_sent_at IS NULL
  AND sc.recovered_at IS NULL
  AND jsonb_array_length(sc.items) > 0;

-- name: GetCartsForEmail3 :many
-- Carts where email_2 was sent 71-73 hours ago, email_3 not yet sent.
SELECT sc.id, sc.shop_id, sc.customer_id, sc.items, sc.coupon_code, sc.total_cents,
       sc.recovery_token, sc.email_1_sent_at, sc.email_2_sent_at, sc.email_3_sent_at,
       sc.recovered_at, sc.created_at, sc.updated_at,
       c.email AS customer_email, c.first_name AS customer_first_name
FROM saved_carts sc
JOIN customers c ON c.id = sc.customer_id
WHERE sc.email_2_sent_at BETWEEN NOW() - INTERVAL '73 hours' AND NOW() - INTERVAL '71 hours'
  AND sc.email_3_sent_at IS NULL
  AND sc.recovered_at IS NULL
  AND jsonb_array_length(sc.items) > 0;

-- name: PurgeOldSavedCarts :exec
DELETE FROM saved_carts
WHERE shop_id = $1 AND updated_at < NOW() - INTERVAL '30 days';

-- name: GetCartRecoveryStats :one
SELECT
    COUNT(*) FILTER (WHERE recovered_at IS NOT NULL)         AS recovered_count,
    COUNT(*) FILTER (WHERE jsonb_array_length(items) > 0)   AS abandoned_count
FROM saved_carts
WHERE shop_id = $1
  AND created_at >= $2;

-- name: HasOrderSinceSavedCart :one
SELECT EXISTS (
    SELECT 1
    FROM orders o
    WHERE o.shop_id = $1
      AND o.customer_id = $2
      AND o.created_at >= $3
      AND (o.status IS NULL OR o.status <> 'cancelled')
) AS has_order;
