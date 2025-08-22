-- ==================== Social Proof ====================

-- name: InsertSocialProofEvent :exec
INSERT INTO social_proof_events (shop_id, product_id, city, country)
VALUES ($1, $2, $3, $4);

-- name: GetRecentSocialProofEvents :many
SELECT
    spe.id,
    spe.shop_id,
    spe.product_id,
    spe.city,
    spe.country,
    spe.created_at,
    p.title AS product_title,
    p.slug  AS product_slug
FROM social_proof_events spe
JOIN products p ON p.id = spe.product_id
WHERE spe.shop_id = $1
ORDER BY spe.created_at DESC
LIMIT $2;

-- name: PurgeStaleSocialProofEvents :exec
DELETE FROM social_proof_events
WHERE shop_id = $1 AND created_at < NOW() - INTERVAL '7 days';
