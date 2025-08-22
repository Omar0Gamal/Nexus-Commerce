-- ==================== Staff Audit Log ====================

-- name: CreateStaffAuditLog :one
INSERT INTO staff_audit_log (
    shop_id, staff_id, action, resource_type, resource_id, metadata, ip_address
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListStaffAuditLog :many
SELECT * FROM staff_audit_log
WHERE shop_id = $1
  AND ($2::uuid IS NULL OR staff_id = $2)
  AND ($3::text  = '' OR resource_type = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;
