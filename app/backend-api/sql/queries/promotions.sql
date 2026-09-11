-- name: GetActiveFlashSales :many
SELECT id, shop_id, name, discount_type, discount_value, starts_at, ends_at, is_active, max_uses, uses_count, created_at
FROM flash_sales
WHERE shop_id = $1 AND is_active = TRUE
  AND NOW() BETWEEN starts_at AND ends_at
ORDER BY starts_at ASC;

-- name: CreateFlashSale :one
INSERT INTO flash_sales (shop_id, name, discount_type, discount_value, starts_at, ends_at, max_uses)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetFlashSaleByID :one
SELECT * FROM flash_sales WHERE id = $1 AND shop_id = $2;

-- name: DeleteFlashSale :exec
UPDATE flash_sales SET is_active = FALSE WHERE id = $1 AND shop_id = $2;

-- name: AddFlashSaleItem :one
INSERT INTO flash_sale_items (flash_sale_id, shop_id, product_id, variant_id, sale_price_cents)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (flash_sale_id, product_id, variant_id) DO UPDATE
SET sale_price_cents = EXCLUDED.sale_price_cents
RETURNING *;

-- name: GetFlashSalePrice :one
SELECT fsi.sale_price_cents
FROM flash_sale_items fsi
JOIN flash_sales fs ON fs.id = fsi.flash_sale_id
WHERE fsi.shop_id = $1
  AND fsi.product_id = $2
  AND (fsi.variant_id = $3 OR fsi.variant_id IS NULL)
  AND fs.is_active = TRUE
  AND NOW() BETWEEN fs.starts_at AND fs.ends_at
ORDER BY fsi.sale_price_cents ASC
LIMIT 1;

-- name: CountActiveFlashSales :one
SELECT COUNT(*) FROM flash_sales
WHERE shop_id = $1 AND is_active = TRUE AND NOW() BETWEEN starts_at AND ends_at;

-- name: GetActiveTieredDiscount :one
SELECT * FROM tiered_discounts
WHERE shop_id = $1 AND applies_to = $2 AND is_active = TRUE
LIMIT 1;

-- name: CreateTieredDiscount :one
INSERT INTO tiered_discounts (shop_id, name, applies_to, target_id, tiers)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetTieredDiscounts :many
SELECT * FROM tiered_discounts WHERE shop_id = $1 AND is_active = TRUE ORDER BY name;

-- name: GetLoyaltyAccount :one
SELECT * FROM loyalty_accounts WHERE shop_id = $1 AND customer_id = $2;

-- name: UpsertLoyaltyAccount :one
INSERT INTO loyalty_accounts (shop_id, customer_id, points_balance, lifetime_points)
VALUES ($1, $2, $3, $3)
ON CONFLICT (shop_id, customer_id) DO UPDATE
SET points_balance  = loyalty_accounts.points_balance + $3,
    lifetime_points = loyalty_accounts.lifetime_points + GREATEST(0, $3),
    updated_at = NOW()
RETURNING *;

-- name: RecordLoyaltyTransaction :one
INSERT INTO loyalty_transactions (shop_id, customer_id, points_delta, reason, reference_id)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetLoyaltyTransactions :many
SELECT * FROM loyalty_transactions
WHERE shop_id = $1 AND customer_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetLoyaltyProgram :one
SELECT * FROM loyalty_programs WHERE shop_id = $1;

-- name: UpsertLoyaltyProgram :one
INSERT INTO loyalty_programs (shop_id, points_per_egp, egp_per_point, min_redeem_points, max_redeem_pct, expiry_months, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (shop_id) DO UPDATE
SET points_per_egp    = EXCLUDED.points_per_egp,
    egp_per_point     = EXCLUDED.egp_per_point,
    min_redeem_points = EXCLUDED.min_redeem_points,
    max_redeem_pct    = EXCLUDED.max_redeem_pct,
    expiry_months     = EXCLUDED.expiry_months,
    is_active         = EXCLUDED.is_active
RETURNING *;

-- name: GetReferralCode :one
SELECT * FROM referral_codes WHERE shop_id = $1 AND customer_id = $2 LIMIT 1;

-- name: CreateReferralCode :one
INSERT INTO referral_codes (shop_id, customer_id, code)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO UPDATE SET uses_count = referral_codes.uses_count
RETURNING *;

-- name: GetReferralCodeByCode :one
SELECT * FROM referral_codes WHERE shop_id = $1 AND code = $2;

-- name: IncrementReferralCodeUses :exec
UPDATE referral_codes SET uses_count = uses_count + 1 WHERE id = $1;

-- name: RecordReferralConversion :one
INSERT INTO referral_conversions (shop_id, referrer_id, referred_id, order_id, referrer_reward_pts, referred_discount_cents)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetMyReferrals :many
SELECT rc.code, rc.uses_count,
       conv.referred_id, conv.referrer_reward_pts, conv.converted_at
FROM referral_codes rc
LEFT JOIN referral_conversions conv ON conv.shop_id = rc.shop_id
    AND conv.referrer_id = rc.customer_id
WHERE rc.shop_id = $1 AND rc.customer_id = $2
ORDER BY conv.converted_at DESC NULLS LAST;

-- name: GetReferralStats :one
SELECT
    COUNT(*)                                                 AS total_referrals,
    COUNT(conv.id)                                           AS total_conversions,
    COALESCE(SUM(conv.referrer_reward_pts), 0)::bigint       AS total_points_awarded
FROM referral_codes rc
LEFT JOIN referral_conversions conv ON conv.shop_id = rc.shop_id
    AND conv.referrer_id = rc.customer_id
WHERE rc.shop_id = $1;

-- name: CountReferralConversionsByReferrer :one
SELECT COUNT(*) FROM referral_conversions
WHERE shop_id = $1 AND referrer_id = $2;

-- name: GetAccountsWithExpiredPoints :many
SELECT la.shop_id, la.customer_id, la.points_balance
FROM loyalty_accounts la
JOIN loyalty_programs lp ON lp.shop_id = la.shop_id
WHERE lp.expiry_months IS NOT NULL
  AND lp.expiry_months > 0
  AND lp.is_active = TRUE
  AND la.points_balance > 0
  AND la.updated_at < NOW() - (lp.expiry_months * INTERVAL '1 month');
