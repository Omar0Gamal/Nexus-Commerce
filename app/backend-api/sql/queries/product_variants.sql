-- ==================== Product Variants ====================

-- name: CreateVariant :one
INSERT INTO product_variants (
    product_id, shop_id, title, sku, price,
    compare_at_price, stock_quantity, options, position
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetVariant :one
SELECT * FROM product_variants
WHERE id = $1 AND shop_id = $2;

-- name: ListVariantsByProduct :many
SELECT * FROM product_variants
WHERE product_id = $1 AND shop_id = $2
ORDER BY position ASC, created_at ASC;

-- name: UpdateVariant :one
UPDATE product_variants SET
    title = COALESCE(sqlc.narg('title'), title),
    sku = COALESCE(sqlc.narg('sku'), sku),
    price = COALESCE(sqlc.narg('price'), price),
    compare_at_price = COALESCE(sqlc.narg('compare_at_price'), compare_at_price),
    stock_quantity = COALESCE(sqlc.narg('stock_quantity'), stock_quantity),
    options = COALESCE(sqlc.narg('options'), options),
    position = COALESCE(sqlc.narg('position'), position)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DecrementVariantStock :one
-- Atomically decrements variant stock only when sufficient units are available.
UPDATE product_variants
SET stock_quantity = stock_quantity - $3
WHERE id = $1
  AND shop_id = $2
  AND stock_quantity >= $3
RETURNING id, stock_quantity;

-- name: DeleteVariant :exec
DELETE FROM product_variants WHERE id = $1 AND shop_id = $2;

-- name: DeleteVariantsByProduct :exec
DELETE FROM product_variants WHERE product_id = $1 AND shop_id = $2;

-- name: CountVariantsByProduct :one
SELECT COUNT(*) FROM product_variants WHERE product_id = $1 AND shop_id = $2;

-- name: GetVariantsByProductIDs :many
SELECT * FROM product_variants
WHERE product_id = ANY($1::uuid[]) AND shop_id = $2
ORDER BY product_id, position ASC, created_at ASC;
