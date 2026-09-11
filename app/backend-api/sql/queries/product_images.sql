-- ==================== Product Images ====================

-- name: CreateImage :one
INSERT INTO product_images (
    product_id, shop_id, variant_id, url, alt_text, position, is_primary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetImage :one
SELECT * FROM product_images
WHERE id = $1 AND shop_id = $2;

-- name: ListImagesByProduct :many
SELECT * FROM product_images
WHERE product_id = $1 AND shop_id = $2
ORDER BY position ASC, created_at ASC;

-- name: ListImagesByVariant :many
SELECT * FROM product_images
WHERE variant_id = $1 AND shop_id = $2
ORDER BY position ASC;

-- name: UpdateImage :one
UPDATE product_images SET
    alt_text = COALESCE(sqlc.narg('alt_text'), alt_text),
    position = COALESCE(sqlc.narg('position'), position),
    is_primary = COALESCE(sqlc.narg('is_primary'), is_primary),
    variant_id = COALESCE(sqlc.narg('variant_id'), variant_id)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteImage :exec
DELETE FROM product_images WHERE id = $1 AND shop_id = $2;

-- name: DeleteImagesByProduct :exec
DELETE FROM product_images WHERE product_id = $1 AND shop_id = $2;

-- name: SetPrimaryImage :exec
UPDATE product_images SET is_primary = (id = $1)
WHERE product_id = $2 AND shop_id = $3;

-- name: GetImagesByProductIDs :many
SELECT * FROM product_images
WHERE product_id = ANY($1::uuid[]) AND shop_id = $2
ORDER BY product_id, position ASC, created_at ASC;
