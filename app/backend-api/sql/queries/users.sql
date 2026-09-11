-- ==================== Users ====================

-- name: CreateUser :one
INSERT INTO users (
    email, password_hash, full_name, phone
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users SET
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    phone = COALESCE(sqlc.narg('phone'), phone),
    is_email_verified = COALESCE(sqlc.narg('is_email_verified'), is_email_verified),
    status = COALESCE(sqlc.narg('status'), status)
WHERE id = @id
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: UpdateUserEmail :exec
UPDATE users SET email = $2, is_email_verified = FALSE WHERE id = $1;

-- name: VerifyUserEmail :exec
UPDATE users SET is_email_verified = TRUE WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
