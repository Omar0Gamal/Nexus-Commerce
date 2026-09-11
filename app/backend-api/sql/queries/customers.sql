-- ==================== Customers ====================

-- name: CreateCustomer :one
INSERT INTO customers (
    shop_id, email, password_hash, first_name, last_name, phone
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetCustomer :one
SELECT * FROM customers
WHERE id = $1 AND shop_id = $2 AND deleted_at IS NULL;

-- name: GetCustomerByEmail :one
SELECT * FROM customers
WHERE email = $1 AND shop_id = $2 AND deleted_at IS NULL;

-- name: ListCustomers :many
SELECT * FROM customers
WHERE shop_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: SearchCustomers :many
SELECT * FROM customers
WHERE shop_id = $1
    AND deleted_at IS NULL
    AND (
        first_name ILIKE '%' || $2 || '%'
        OR last_name ILIKE '%' || $2 || '%'
        OR email ILIKE '%' || $2 || '%'
    )
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateCustomer :one
UPDATE customers SET
    email = COALESCE(sqlc.narg('email'), email),
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name = COALESCE(sqlc.narg('last_name'), last_name),
    phone = COALESCE(sqlc.narg('phone'), phone)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers WHERE id = $1 AND shop_id = $2;

-- name: CountCustomers :one
SELECT COUNT(*) FROM customers WHERE shop_id = $1 AND deleted_at IS NULL;

-- name: UpdateCustomerPassword :exec
UPDATE customers SET password_hash = $3 WHERE id = $1 AND shop_id = $2;

-- name: SoftDeleteCustomer :exec
UPDATE customers SET deleted_at = NOW()
WHERE id = $1 AND shop_id = $2 AND deleted_at IS NULL;

-- name: RestoreCustomer :one
UPDATE customers SET deleted_at = NULL
WHERE id = $1 AND shop_id = $2 AND deleted_at IS NOT NULL
RETURNING *;

-- name: ListDeletedCustomers :many
SELECT * FROM customers
WHERE shop_id = $1 AND deleted_at IS NOT NULL
ORDER BY deleted_at DESC;

-- name: GetAtRiskCustomers :many
WITH CustomerIntervals AS (
    SELECT 
        customer_id,
        COUNT(id) as order_count,
        MAX(created_at) as last_order_date,
        EXTRACT(EPOCH FROM (MAX(created_at) - MIN(created_at))) / 86400.0 / NULLIF(COUNT(id) - 1, 0) as avg_interval_days
    FROM orders
    WHERE shop_id = sqlc.arg('shop_id') AND status != 'cancelled'
    GROUP BY customer_id
    HAVING COUNT(id) >= 3
)
SELECT 
    customer_id,
    avg_interval_days::float8 as avg_order_interval_days,
    (EXTRACT(EPOCH FROM (NOW() - last_order_date)) / 86400.0)::float8 as days_since_last_order
FROM CustomerIntervals
WHERE 
    avg_interval_days IS NOT NULL 
    AND EXTRACT(EPOCH FROM (NOW() - last_order_date)) / 86400.0 > (avg_interval_days * 2)
ORDER BY days_since_last_order DESC
LIMIT sqlc.arg('limit_count');
