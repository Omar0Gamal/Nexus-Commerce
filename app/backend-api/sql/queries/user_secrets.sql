-- ==================== User Secrets ====================

-- name: CreateUserSecrets :one
INSERT INTO user_secrets (
    user_id, totp_secret, backup_codes, is_2fa_enabled, recovery_email
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUserSecrets :one
SELECT * FROM user_secrets WHERE user_id = $1;

-- name: UpdateUserSecrets :one
UPDATE user_secrets SET
    totp_secret = COALESCE(sqlc.narg('totp_secret'), totp_secret),
    backup_codes = COALESCE(sqlc.narg('backup_codes'), backup_codes),
    is_2fa_enabled = COALESCE(sqlc.narg('is_2fa_enabled'), is_2fa_enabled),
    recovery_email = COALESCE(sqlc.narg('recovery_email'), recovery_email),
    updated_at = NOW()
WHERE user_id = @user_id
RETURNING *;

-- name: Enable2FA :exec
UPDATE user_secrets SET
    totp_secret = $2,
    backup_codes = $3,
    is_2fa_enabled = TRUE,
    updated_at = NOW()
WHERE user_id = $1;

-- name: Disable2FA :exec
UPDATE user_secrets SET
    totp_secret = NULL,
    backup_codes = '[]',
    is_2fa_enabled = FALSE,
    updated_at = NOW()
WHERE user_id = $1;

-- name: DeleteUserSecrets :exec
DELETE FROM user_secrets WHERE user_id = $1;
