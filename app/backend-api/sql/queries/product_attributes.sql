-- name: UpsertProductAttribute :one
INSERT INTO product_attributes (shop_id, product_id, key, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (product_id, key) DO UPDATE SET value = EXCLUDED.value
RETURNING *;

-- name: DeleteProductAttribute :exec
DELETE FROM product_attributes
WHERE product_id = $1 AND key = $2;

-- name: ListProductAttributes :many
SELECT * FROM product_attributes
WHERE product_id = $1
ORDER BY key;

-- name: GetAttributesByShop :many
SELECT DISTINCT key, value
FROM product_attributes
WHERE shop_id = $1
ORDER BY key, value;

-- name: DeleteProductAttributes :exec
DELETE FROM product_attributes WHERE product_id = $1;
