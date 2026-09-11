-- ==================== Support Messages ====================

-- name: CreateSupportMessage :one
INSERT INTO support_messages (
    ticket_id, sender_type, staff_id, customer_id,
    message_body, attachments, is_internal_note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetSupportMessage :one
SELECT * FROM support_messages WHERE id = $1;

-- name: ListSupportMessages :many
SELECT * FROM support_messages
WHERE ticket_id = $1
ORDER BY created_at ASC;

-- name: ListPublicSupportMessages :many
SELECT * FROM support_messages
WHERE ticket_id = $1 AND is_internal_note = FALSE
ORDER BY created_at ASC;

-- name: ListInternalNotes :many
SELECT * FROM support_messages
WHERE ticket_id = $1 AND is_internal_note = TRUE
ORDER BY created_at ASC;

-- name: MarkMessageAsRead :exec
UPDATE support_messages SET read_at = NOW()
WHERE id = $1 AND read_at IS NULL;

-- name: MarkAllMessagesAsRead :exec
UPDATE support_messages SET read_at = NOW()
WHERE ticket_id = $1 AND read_at IS NULL AND sender_type = $2;

-- name: CountUnreadMessages :one
SELECT COUNT(*) FROM support_messages
WHERE ticket_id = $1 AND read_at IS NULL AND sender_type = $2;

-- name: DeleteSupportMessage :exec
DELETE FROM support_messages WHERE id = $1;
