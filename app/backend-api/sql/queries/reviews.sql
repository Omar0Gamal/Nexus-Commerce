-- ==================== Product Reviews ====================

-- name: CreateReview :one
INSERT INTO product_reviews
    (shop_id, product_id, customer_id, order_id, rating, title, body, status, is_verified_purchase)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetReview :one
SELECT * FROM product_reviews
WHERE id = $1 AND shop_id = $2;

-- name: GetReviewsByProduct :many
SELECT
    r.id, r.shop_id, r.product_id, r.customer_id, r.rating,
    r.title, r.body, r.status, r.is_verified_purchase, r.helpful_count,
    r.created_at, r.updated_at,
    c.first_name, c.last_name
FROM product_reviews r
JOIN customers c ON c.id = r.customer_id
WHERE r.shop_id = $1 AND r.product_id = $2 AND r.status = 'approved'
ORDER BY r.helpful_count DESC, r.created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetReviewSummary :one
SELECT
    COUNT(*)::INTEGER                                   AS total_reviews,
    COALESCE(AVG(rating), 0)::NUMERIC(3,2)             AS avg_rating,
    COUNT(*) FILTER (WHERE rating = 5)::INTEGER         AS five_star,
    COUNT(*) FILTER (WHERE rating = 4)::INTEGER         AS four_star,
    COUNT(*) FILTER (WHERE rating = 3)::INTEGER         AS three_star,
    COUNT(*) FILTER (WHERE rating = 2)::INTEGER         AS two_star,
    COUNT(*) FILTER (WHERE rating = 1)::INTEGER         AS one_star
FROM product_reviews
WHERE shop_id = $1 AND product_id = $2 AND status = 'approved';

-- name: GetPendingReviews :many
SELECT
    r.id, r.shop_id, r.product_id, r.customer_id, r.rating,
    r.title, r.body, r.status, r.is_verified_purchase, r.helpful_count,
    r.created_at, r.updated_at,
    c.first_name,
    p.title AS product_title
FROM product_reviews r
JOIN customers c ON c.id = r.customer_id
JOIN products p ON p.id = r.product_id
WHERE r.shop_id = $1 AND r.status = 'pending'
ORDER BY r.created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdateReviewStatus :one
UPDATE product_reviews
SET status = $3, updated_at = NOW()
WHERE shop_id = $1 AND id = $2
RETURNING *;

-- name: IncrementHelpfulCount :exec
UPDATE product_reviews
SET helpful_count = helpful_count + 1
WHERE id = $1 AND shop_id = $2;

-- name: SyncProductRatingCache :exec
UPDATE products
SET review_count = sub.cnt,
    avg_rating   = sub.avg
FROM (
    SELECT
        COUNT(*)::INTEGER               AS cnt,
        COALESCE(AVG(rating), 0)::NUMERIC(3,2) AS avg
    FROM product_reviews
    WHERE product_reviews.shop_id = $1 AND product_reviews.product_id = $2 AND status = 'approved'
) sub
WHERE products.id = $2 AND products.shop_id = $1;

-- name: CustomerHasPurchasedProduct :one
SELECT COUNT(*) > 0 AS purchased
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE o.shop_id = $1
  AND o.customer_id = $2
  AND oi.product_id = $3
  AND o.status = 'completed';
