-- ==================== Orders ====================

-- name: CreateOrder :one
INSERT INTO orders (
    shop_id, customer_id, coupon_id, order_number, total_price,
    discount_amount, shipping_fee, status, shipping_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1 AND shop_id = $2;

-- name: GetOrderByNumber :one
SELECT * FROM orders
WHERE order_number = $1 AND shop_id = $2;

-- name: ListOrders :many
SELECT * FROM orders
WHERE shop_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersByStatus :many
SELECT * FROM orders
WHERE shop_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListOrdersByCustomer :many
SELECT * FROM orders
WHERE shop_id = $1 AND customer_id = $2
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :one
UPDATE orders SET status = $2
WHERE id = $1 AND shop_id = $3
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = $1 AND shop_id = $2;

-- name: CountOrders :one
SELECT COUNT(*) FROM orders WHERE shop_id = $1;

-- name: CountOrdersByStatus :one
SELECT COUNT(*) FROM orders WHERE shop_id = $1 AND status = $2;

-- name: GetNextOrderNumber :one
SELECT COALESCE(MAX(order_number), 0) + 1 AS next_order_number
FROM orders WHERE shop_id = $1;

-- name: GetTotalRevenue :one
SELECT COALESCE(SUM(total_price), 0)::numeric FROM orders WHERE shop_id = $1;

-- name: GetCustomerOrderSummary :one
SELECT
    COUNT(*)                                  AS order_count,
    COALESCE(SUM(total_price), 0)::numeric    AS total_spent,
    MAX(created_at)                           AS last_order_at
FROM orders
WHERE shop_id = $1 AND customer_id = $2;
