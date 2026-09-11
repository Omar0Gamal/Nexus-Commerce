-- ==================== Invitations ====================

-- name: CreateInvitation :one
INSERT INTO invitations (
    email, token, type, payload, invited_by_user_id, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetInvitation :one
SELECT * FROM invitations WHERE id = $1;

-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token = $1;

-- name: ListInvitationsByEmail :many
SELECT * FROM invitations
WHERE email = $1
ORDER BY created_at DESC;

-- name: ListInvitationsByInviter :many
SELECT * FROM invitations
WHERE invited_by_user_id = $1
ORDER BY created_at DESC;

-- name: DeleteInvitation :exec
DELETE FROM invitations WHERE id = $1;

-- name: DeleteExpiredInvitations :exec
DELETE FROM invitations WHERE expires_at < NOW();
