-- name: CreatePlatformTicket :one
INSERT INTO platform_tickets (shop_id, subject, status, priority)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreatePlatformTicketMessage :one
INSERT INTO platform_ticket_messages (ticket_id, sender_type, message)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPlatformTicketsByShop :many
SELECT * FROM platform_tickets
WHERE shop_id = $1
ORDER BY created_at DESC;

-- name: GetAllPlatformTickets :many
SELECT * FROM platform_tickets
ORDER BY priority DESC, created_at DESC;

-- name: GetPlatformTicketByID :one
SELECT * FROM platform_tickets
WHERE id = $1;

-- name: GetPlatformTicketMessages :many
SELECT * FROM platform_ticket_messages
WHERE ticket_id = $1
ORDER BY created_at ASC;

-- name: UpdatePlatformTicketStatus :one
UPDATE platform_tickets
SET status = $2
WHERE id = $1
RETURNING *;
