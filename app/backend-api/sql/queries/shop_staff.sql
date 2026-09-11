-- ==================== Shop Staff ====================

-- name: CreateShopStaff :one
INSERT INTO shop_staff (
    shop_id, user_id, role_id, is_owner
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetShopStaff :one
SELECT * FROM shop_staff WHERE id = $1;

-- name: GetShopStaffByUserAndShop :one
SELECT * FROM shop_staff
WHERE shop_id = $1 AND user_id = $2;

-- name: ListShopStaff :many
SELECT * FROM shop_staff
WHERE shop_id = $1
ORDER BY joined_at ASC;

-- name: ListStaffByUser :many
SELECT * FROM shop_staff
WHERE user_id = $1
ORDER BY joined_at ASC;

-- name: GetShopOwner :one
SELECT * FROM shop_staff
WHERE shop_id = $1 AND is_owner = TRUE;

-- name: UpdateShopStaffRole :one
UPDATE shop_staff SET role_id = $2
WHERE id = $1
RETURNING *;

-- name: DeleteShopStaff :exec
DELETE FROM shop_staff WHERE id = $1;

-- name: CountShopStaff :one
SELECT COUNT(*) FROM shop_staff WHERE shop_id = $1;

-- name: CountStaffByRole :one
SELECT COUNT(*) FROM shop_staff WHERE role_id = $1;

-- name: ListStaffWithUsers :many
SELECT ss.id, ss.user_id, COALESCE(ss.is_owner, false) AS is_owner, ss.joined_at,
       u.email, u.full_name, u.phone
FROM shop_staff ss
JOIN users u ON u.id = ss.user_id
WHERE ss.shop_id = $1
ORDER BY ss.joined_at ASC;

-- name: DeleteShopStaffProtected :exec
DELETE FROM shop_staff
WHERE id = $1 AND shop_id = $2 AND COALESCE(is_owner, false) = false;

-- name: ListShopsForUser :many
SELECT
    s.id         AS shop_id,
    s.name       AS shop_name,
    s.subdomain,
    sr.name      AS role_name,
    sr.permissions AS role_permissions,
    COALESCE(ss.is_owner, false) AS is_owner,
    ss.joined_at
FROM shop_staff ss
JOIN shops s  ON s.id  = ss.shop_id
LEFT JOIN shop_roles sr ON sr.id = ss.role_id
WHERE ss.user_id = $1
ORDER BY ss.joined_at ASC;

-- name: ListStaffMembersDetailed :many
SELECT
    ss.id,
    ss.shop_id,
    ss.user_id,
    ss.role_id,
    COALESCE(ss.is_owner, false) AS is_owner,
    ss.joined_at,
    u.email,
    COALESCE(u.full_name, '') AS full_name,
    COALESCE(u.phone, '') AS phone,
    u.status AS user_status,
    COALESCE(sr.name, '') AS role_name
FROM shop_staff ss
JOIN users u ON u.id = ss.user_id
LEFT JOIN shop_roles sr ON sr.id = ss.role_id
WHERE ss.shop_id = $1
ORDER BY ss.joined_at ASC;

-- name: GetStaffMemberDetailed :one
SELECT
    ss.id,
    ss.shop_id,
    ss.user_id,
    ss.role_id,
    COALESCE(ss.is_owner, false) AS is_owner,
    ss.joined_at,
    u.email,
    COALESCE(u.full_name, '') AS full_name,
    COALESCE(u.phone, '') AS phone,
    u.status AS user_status,
    COALESCE(sr.name, '') AS role_name
FROM shop_staff ss
JOIN users u ON u.id = ss.user_id
LEFT JOIN shop_roles sr ON sr.id = ss.role_id
WHERE ss.id = $1 AND ss.shop_id = $2;

-- name: UpdateShopStaffRoleForShop :one
UPDATE shop_staff SET role_id = $3
WHERE id = $1 AND shop_id = $2
RETURNING *;

