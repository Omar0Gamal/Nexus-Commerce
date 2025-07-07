-- ==================== Shops ====================

-- name: CreateShop :one
INSERT INTO shops (
    plan_id, owner_user_id, name, subdomain,
    custom_domain, status, currency, timezone
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetShop :one
SELECT * FROM shops WHERE id = $1;

-- name: GetShopBySubdomain :one
SELECT * FROM shops WHERE subdomain = $1;

-- name: GetShopByCustomDomain :one
SELECT * FROM shops WHERE custom_domain = $1;

-- name: ListShops :many
SELECT * FROM shops
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminListShops :many
SELECT 
    s.id, s.name, s.subdomain, s.custom_domain, s.status, s.currency, s.timezone, s.owner_user_id, s.plan_id, s.created_at,
    p.name AS plan_name
FROM shops s
JOIN plans p ON s.plan_id = p.id
ORDER BY
  CASE WHEN sqlc.arg('sort')::text = 'name' AND sqlc.arg('order')::text = 'asc' THEN s.name END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'name' AND sqlc.arg('order')::text = 'desc' THEN s.name END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'subdomain' AND sqlc.arg('order')::text = 'asc' THEN s.subdomain END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'subdomain' AND sqlc.arg('order')::text = 'desc' THEN s.subdomain END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'plan_name' AND sqlc.arg('order')::text = 'asc' THEN p.name END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'plan_name' AND sqlc.arg('order')::text = 'desc' THEN p.name END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'status' AND sqlc.arg('order')::text = 'asc' THEN s.status::text END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'status' AND sqlc.arg('order')::text = 'desc' THEN s.status::text END DESC,
  CASE WHEN sqlc.arg('sort')::text = 'created_at' AND sqlc.arg('order')::text = 'asc' THEN s.created_at END ASC,
  CASE WHEN sqlc.arg('sort')::text = 'created_at' AND sqlc.arg('order')::text = 'desc' THEN s.created_at END DESC,
  s.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListShopsByOwner :many
SELECT * FROM shops WHERE owner_user_id = $1 ORDER BY created_at DESC;

-- name: ListShopsByStatus :many
SELECT * FROM shops WHERE status = $1 ORDER BY created_at DESC;

-- name: UpdateShop :one
UPDATE shops SET
    name = COALESCE(sqlc.narg('name'), name),
    subdomain = COALESCE(sqlc.narg('subdomain'), subdomain),
    custom_domain = COALESCE(sqlc.narg('custom_domain'), custom_domain),
    status = COALESCE(sqlc.narg('status'), status),
    currency = COALESCE(sqlc.narg('currency'), currency),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    plan_id = COALESCE(sqlc.narg('plan_id'), plan_id),
    current_period_end = COALESCE(sqlc.narg('current_period_end'), current_period_end),
    is_overdue = COALESCE(sqlc.narg('is_overdue'), is_overdue)
WHERE id = @id
RETURNING *;

-- name: UpdateShopStatus :exec
UPDATE shops SET status = $2 WHERE id = $1;

-- name: UpdateShopBilling :exec
UPDATE shops SET
    current_period_end = $2,
    is_overdue = $3
WHERE id = $1;

-- name: DeleteShop :exec
DELETE FROM shops WHERE id = $1;

-- name: CountShops :one
SELECT COUNT(*) FROM shops;

-- name: CountShopsByStatus :one
SELECT COUNT(*) FROM shops WHERE status = $1;

-- name: UpdateShopSEO :exec
UPDATE shops
SET seo_title = $2, seo_description = $3, favicon_url = $4
WHERE id = $1;

-- name: GetShopPlanName :one
SELECT p.name
FROM shops s
JOIN plans p ON p.id = s.plan_id
WHERE s.id = $1;

-- name: GetShopNotificationPrefs :one
SELECT COALESCE(notification_prefs, '{"new_orders":true,"low_stock":true,"customer_reviews":true}'::jsonb)
FROM shops
WHERE id = $1;

-- name: UpdateShopNotificationPrefs :exec
UPDATE shops
SET notification_prefs = $2
WHERE id = $1;

-- name: GetBillingSummary :one
SELECT
    COUNT(*) FILTER (WHERE s.status = 'active')           AS active_shops,
    COUNT(*) FILTER (WHERE s.status = 'suspended')        AS suspended_shops,
    COUNT(*) FILTER (WHERE s.created_at >= NOW() - INTERVAL '30 days') AS new_last_30d,
    COALESCE(SUM(p.monthly_price) FILTER (WHERE s.status = 'active'), 0) AS mrr
FROM shops s
JOIN plans p ON p.id = s.plan_id;

-- name: GetPlanDistribution :many
SELECT p.name AS plan_name, COUNT(s.id) AS shop_count
FROM plans p
LEFT JOIN shops s ON s.plan_id = p.id AND s.status = 'active'
GROUP BY p.name
ORDER BY shop_count DESC;

-- name: CountNewShops :one
SELECT COUNT(*) FROM shops WHERE created_at >= $1;

