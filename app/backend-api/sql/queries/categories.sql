-- ==================== Categories ====================

-- name: CreateCategory :one
INSERT INTO categories (
    shop_id, parent_id, name, slug
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetCategory :one
SELECT * FROM categories
WHERE id = $1 AND shop_id = $2;

-- name: GetCategoryBySlug :one
SELECT * FROM categories
WHERE slug = $1 AND shop_id = $2;

-- name: ListCategories :many
SELECT * FROM categories
WHERE shop_id = $1
ORDER BY name ASC;

-- name: ListRootCategories :many
SELECT * FROM categories
WHERE shop_id = $1 AND parent_id IS NULL
ORDER BY name ASC;

-- name: ListSubcategories :many
SELECT * FROM categories
WHERE shop_id = $1 AND parent_id = $2
ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE categories SET
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    parent_id = COALESCE(sqlc.narg('parent_id'), parent_id)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1 AND shop_id = $2;

-- name: UpdateCategorySEO :exec
UPDATE categories
SET meta_description = $3
WHERE id = $1 AND shop_id = $2;

-- name: CountCategories :one
SELECT COUNT(*) FROM categories WHERE shop_id = $1;
