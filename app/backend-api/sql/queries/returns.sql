-- ==================== Order Returns ====================

-- name: CreateReturn :one
INSERT INTO order_returns (shop_id, order_id, customer_id, reason, status)
VALUES ($1, $2, $3, $4, 'requested')
RETURNING *;

-- name: GetReturn :one
SELECT * FROM order_returns
WHERE id = $1 AND shop_id = $2;

-- name: ListReturns :many
SELECT * FROM order_returns
WHERE shop_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListReturnsByOrder :many
SELECT * FROM order_returns
WHERE order_id = $1 AND shop_id = $2
ORDER BY created_at DESC;

-- name: UpdateReturnStatus :one
UPDATE order_returns
SET status     = $3,
    notes      = COALESCE($4, notes),
    updated_at = NOW()
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: SetReturnRefunded :one
UPDATE order_returns
SET status           = 'refunded',
    paymob_refund_id = $3,
    refund_amount    = $4,
    updated_at       = NOW()
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: CountReturns :one
SELECT COUNT(*) FROM order_returns
WHERE shop_id = $1;
