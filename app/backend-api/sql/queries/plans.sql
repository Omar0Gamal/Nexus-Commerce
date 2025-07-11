-- ==================== Plans ====================

-- name: CreatePlan :one
INSERT INTO plans (
    name, monthly_price, max_products, max_staff_accounts,
    max_storage_mb, transaction_fee_percent, features
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = $1;

-- name: GetPlanByName :one
SELECT * FROM plans WHERE name = $1;

-- name: ListPlans :many
SELECT * FROM plans ORDER BY monthly_price ASC;

-- name: UpdatePlan :one
UPDATE plans SET
    name = COALESCE(sqlc.narg('name'), name),
    monthly_price = COALESCE(sqlc.narg('monthly_price'), monthly_price),
    max_products = COALESCE(sqlc.narg('max_products'), max_products),
    max_staff_accounts = COALESCE(sqlc.narg('max_staff_accounts'), max_staff_accounts),
    max_storage_mb = COALESCE(sqlc.narg('max_storage_mb'), max_storage_mb),
    transaction_fee_percent = COALESCE(sqlc.narg('transaction_fee_percent'), transaction_fee_percent),
    features = COALESCE(sqlc.narg('features'), features)
WHERE id = @id
RETURNING *;

-- name: DeletePlan :exec
DELETE FROM plans WHERE id = $1;
