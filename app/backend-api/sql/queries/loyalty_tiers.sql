-- ==================== Loyalty Tiers ====================

-- name: CreateLoyaltyTier :one
INSERT INTO loyalty_tiers (shop_id, name, min_lifetime_points, benefits, badge_label, badge_color, position)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListLoyaltyTiers :many
SELECT * FROM loyalty_tiers
WHERE shop_id = $1
ORDER BY position ASC;

-- name: GetLoyaltyTier :one
SELECT * FROM loyalty_tiers WHERE id = $1 AND shop_id = $2;

-- name: UpdateLoyaltyTier :one
UPDATE loyalty_tiers SET
    name                = COALESCE(sqlc.narg('name'), name),
    min_lifetime_points = COALESCE(sqlc.narg('min_lifetime_points'), min_lifetime_points),
    benefits            = COALESCE(sqlc.narg('benefits'), benefits),
    badge_label         = COALESCE(sqlc.narg('badge_label'), badge_label),
    badge_color         = COALESCE(sqlc.narg('badge_color'), badge_color),
    position            = COALESCE(sqlc.narg('position'), position)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteLoyaltyTier :exec
DELETE FROM loyalty_tiers WHERE id = $1 AND shop_id = $2;

-- name: GetCustomerCurrentTier :one
SELECT lt.*
FROM loyalty_tiers lt
JOIN loyalty_accounts la ON la.shop_id = lt.shop_id
WHERE la.shop_id = $1
  AND la.customer_id = $2
  AND lt.min_lifetime_points <= la.lifetime_points
ORDER BY lt.min_lifetime_points DESC
LIMIT 1;

-- name: GetNextTier :one
SELECT lt.*
FROM loyalty_tiers lt
JOIN loyalty_accounts la ON la.shop_id = lt.shop_id
WHERE la.shop_id = $1
  AND la.customer_id = $2
  AND lt.min_lifetime_points > la.lifetime_points
ORDER BY lt.min_lifetime_points ASC
LIMIT 1;

-- name: GetLoyaltyLeaderboard :many
SELECT la.customer_id, la.points_balance, la.lifetime_points
FROM loyalty_accounts la
WHERE la.shop_id = $1
ORDER BY la.points_balance DESC
LIMIT $2;
