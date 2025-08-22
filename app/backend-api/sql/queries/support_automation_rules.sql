-- ==================== Support Automation Rules ====================

-- name: CreateSupportAutomationRule :one
INSERT INTO support_automation_rules (
    shop_id, name, is_active, conditions,
    action_assign_staff_id, action_assign_priority,
    action_auto_reply_text, priority_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetSupportAutomationRule :one
SELECT * FROM support_automation_rules
WHERE id = $1 AND shop_id = $2;

-- name: ListSupportAutomationRules :many
SELECT * FROM support_automation_rules
WHERE shop_id = $1
ORDER BY priority_order ASC;

-- name: ListActiveSupportAutomationRules :many
SELECT * FROM support_automation_rules
WHERE shop_id = $1 AND is_active = TRUE
ORDER BY priority_order ASC;

-- name: UpdateSupportAutomationRule :one
UPDATE support_automation_rules SET
    name = COALESCE(sqlc.narg('name'), name),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    conditions = COALESCE(sqlc.narg('conditions'), conditions),
    action_assign_staff_id = COALESCE(sqlc.narg('action_assign_staff_id'), action_assign_staff_id),
    action_assign_priority = COALESCE(sqlc.narg('action_assign_priority'), action_assign_priority),
    action_auto_reply_text = COALESCE(sqlc.narg('action_auto_reply_text'), action_auto_reply_text),
    priority_order = COALESCE(sqlc.narg('priority_order'), priority_order)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: ToggleSupportAutomationRule :one
UPDATE support_automation_rules SET is_active = NOT is_active
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: DeleteSupportAutomationRule :exec
DELETE FROM support_automation_rules WHERE id = $1 AND shop_id = $2;
