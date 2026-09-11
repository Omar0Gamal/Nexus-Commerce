-- ==================== Shop Roles ====================

-- name: CreateShopRole :one
INSERT INTO shop_roles (
    shop_id, parent_role_id, name, permissions, is_system_role
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetShopRole :one
SELECT * FROM shop_roles WHERE id = $1;

-- name: ListShopRoles :many
SELECT * FROM shop_roles
WHERE shop_id = $1
ORDER BY name ASC;

-- name: ListSystemRoles :many
SELECT * FROM shop_roles
WHERE shop_id = $1 AND is_system_role = TRUE
ORDER BY name ASC;

-- name: UpdateShopRole :one
UPDATE shop_roles SET
    name = COALESCE(sqlc.narg('name'), name),
    permissions = COALESCE(sqlc.narg('permissions'), permissions),
    parent_role_id = COALESCE(sqlc.narg('parent_role_id'), parent_role_id)
WHERE id = @id AND is_system_role = FALSE
RETURNING *;

-- name: DeleteShopRole :exec
DELETE FROM shop_roles WHERE id = $1 AND is_system_role = FALSE;
