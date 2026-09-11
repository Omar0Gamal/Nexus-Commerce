-- ==================== Platform Admins ====================

-- name: CreatePlatformAdmin :one
INSERT INTO platform_admins (
    email, password_hash, full_name
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPlatformAdmin :one
SELECT * FROM platform_admins WHERE id = $1;

-- name: GetPlatformAdminByEmail :one
SELECT * FROM platform_admins WHERE email = $1;

-- name: ListPlatformAdmins :many
SELECT * FROM platform_admins ORDER BY created_at DESC;

-- name: UpdatePlatformAdmin :one
UPDATE platform_admins SET
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    is_2fa_enabled = COALESCE(sqlc.narg('is_2fa_enabled'), is_2fa_enabled)
WHERE id = @id
RETURNING *;

-- name: UpdatePlatformAdminPassword :exec
UPDATE platform_admins SET password_hash = $2 WHERE id = $1;

-- name: DeletePlatformAdmin :exec
DELETE FROM platform_admins WHERE id = $1;
