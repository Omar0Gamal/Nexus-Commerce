-- ==================== Order Items ====================

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id, product_id, variant_id, name, price, quantity
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetOrderItem :one
SELECT * FROM order_items WHERE id = $1;

-- name: ListOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY name ASC;

-- name: UpdateOrderItem :one
UPDATE order_items SET
    quantity = COALESCE(sqlc.narg('quantity'), quantity),
    price = COALESCE(sqlc.narg('price'), price)
WHERE id = @id
RETURNING *;

-- name: DeleteOrderItem :exec
DELETE FROM order_items WHERE id = $1;

-- name: DeleteOrderItemsByOrder :exec
DELETE FROM order_items WHERE order_id = $1;
