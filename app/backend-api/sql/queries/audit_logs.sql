-- ==================== Audit Logs ====================

-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    shop_id, actor_user_id, actor_name, ip_address,
    action, resource_type, resource_id, changes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetAuditLog :one
SELECT * FROM audit_logs WHERE id = $1;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE shop_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByActor :many
SELECT * FROM audit_logs
WHERE shop_id = $1 AND actor_user_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAuditLogsByResource :many
SELECT * FROM audit_logs
WHERE shop_id = $1 AND resource_type = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAuditLogsByResourceID :many
SELECT * FROM audit_logs
WHERE shop_id = $1 AND resource_type = $2 AND resource_id = $3
ORDER BY created_at DESC;

-- name: ListAuditLogsByAction :many
SELECT * FROM audit_logs
WHERE shop_id = $1 AND action = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: DeleteAuditLogsBefore :exec
DELETE FROM audit_logs
WHERE shop_id = $1 AND created_at < $2;
