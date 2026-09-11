-- ==================== Support Tickets ====================

-- name: CreateSupportTicket :one
INSERT INTO support_tickets (
    shop_id, customer_id, guest_email, subject,
    status, priority, category, assigned_staff_id, related_order_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetSupportTicket :one
SELECT * FROM support_tickets
WHERE id = $1 AND shop_id = $2;

-- name: ListSupportTickets :many
SELECT * FROM support_tickets
WHERE shop_id = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSupportTicketsByStatus :many
SELECT * FROM support_tickets
WHERE shop_id = $1 AND status = $2
ORDER BY updated_at DESC
LIMIT $3 OFFSET $4;

-- name: ListSupportTicketsByPriority :many
SELECT * FROM support_tickets
WHERE shop_id = $1 AND priority = $2
ORDER BY updated_at DESC;

-- name: ListSupportTicketsByCustomer :many
SELECT * FROM support_tickets
WHERE shop_id = $1 AND customer_id = $2
ORDER BY updated_at DESC;

-- name: ListSupportTicketsByStaff :many
SELECT * FROM support_tickets
WHERE assigned_staff_id = $1
ORDER BY updated_at DESC;

-- name: ListUnassignedTickets :many
SELECT * FROM support_tickets
WHERE shop_id = $1 AND assigned_staff_id IS NULL AND status = 'open'
ORDER BY created_at ASC;

-- name: UpdateSupportTicket :one
UPDATE support_tickets SET
    status = COALESCE(sqlc.narg('status'), status),
    priority = COALESCE(sqlc.narg('priority'), priority),
    category = COALESCE(sqlc.narg('category'), category),
    assigned_staff_id = COALESCE(sqlc.narg('assigned_staff_id'), assigned_staff_id),
    updated_at = NOW()
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: AssignSupportTicket :one
UPDATE support_tickets SET
    assigned_staff_id = $2,
    status = 'pending_staff',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CloseSupportTicket :one
UPDATE support_tickets SET
    status = 'closed',
    updated_at = NOW()
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: ResolveSupportTicket :one
UPDATE support_tickets SET
    status = 'resolved',
    updated_at = NOW()
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: DeleteSupportTicket :exec
DELETE FROM support_tickets WHERE id = $1 AND shop_id = $2;

-- name: CountSupportTickets :one
SELECT COUNT(*) FROM support_tickets WHERE shop_id = $1;

-- name: CountSupportTicketsByStatus :one
SELECT COUNT(*) FROM support_tickets WHERE shop_id = $1 AND status = $2;
