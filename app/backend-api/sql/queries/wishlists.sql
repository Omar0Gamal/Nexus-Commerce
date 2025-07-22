-- ==================== Wishlists ====================

-- name: AddToWishlist :one
INSERT INTO wishlists (shop_id, customer_id, product_id, variant_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (shop_id, customer_id, product_id) DO NOTHING
RETURNING *;

-- name: RemoveFromWishlist :exec
DELETE FROM wishlists
WHERE shop_id = $1 AND customer_id = $2 AND product_id = $3;

-- name: GetWishlist :many
SELECT
    w.id,
    w.shop_id,
    w.customer_id,
    w.product_id,
    w.variant_id,
    w.created_at,
    p.title       AS product_title,
    p.slug        AS product_slug,
    p.price       AS product_price,
    p.stock_quantity AS product_stock
FROM wishlists w
JOIN products p ON p.id = w.product_id
WHERE w.shop_id = $1 AND w.customer_id = $2
ORDER BY w.created_at DESC;

-- name: IsWishlisted :one
SELECT EXISTS (
    SELECT 1 FROM wishlists
    WHERE shop_id = $1 AND customer_id = $2 AND product_id = $3
) AS wishlisted;

-- name: GetWishlistedCustomersForProduct :many
SELECT DISTINCT w.customer_id, c.email, c.first_name
FROM wishlists w
JOIN customers c ON c.id = w.customer_id
WHERE w.shop_id = $1 AND w.product_id = $2;

-- name: GetMostWishlistedProducts :many
SELECT product_id, COUNT(*) AS wishlist_count
FROM wishlists
WHERE shop_id = $1
GROUP BY product_id
ORDER BY wishlist_count DESC
LIMIT $2;
