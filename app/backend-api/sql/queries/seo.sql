-- ==================== SEO ====================

-- ── URL Redirects ─────────────────────────────────────────────────────────────

-- name: CreateRedirect :one
INSERT INTO url_redirects (shop_id, from_path, to_path, status_code)
VALUES ($1, $2, $3, $4)
ON CONFLICT (shop_id, from_path) DO UPDATE
    SET to_path = EXCLUDED.to_path, status_code = EXCLUDED.status_code
RETURNING *;

-- name: GetRedirect :one
SELECT * FROM url_redirects
WHERE shop_id = $1 AND from_path = $2;

-- name: ListRedirects :many
SELECT * FROM url_redirects
WHERE shop_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountRedirects :one
SELECT COUNT(*) FROM url_redirects
WHERE shop_id = $1;

-- name: DeleteRedirect :exec
DELETE FROM url_redirects
WHERE id = $1 AND shop_id = $2;

-- name: IncrementRedirectHitCount :exec
UPDATE url_redirects
SET hit_count = hit_count + 1
WHERE shop_id = $1 AND from_path = $2;

-- ── SEO Scoring ───────────────────────────────────────────────────────────────

-- name: UpdateProductSEOScore :exec
UPDATE products
SET seo_score = $3
WHERE id = $1 AND shop_id = $2;

-- name: UpdateProductStructuredData :exec
UPDATE products
SET structured_data = $3
WHERE id = $1 AND shop_id = $2;

-- name: GetProductsForSEOAudit :many
SELECT id, title, slug, seo_title, seo_description, description,
       review_count, price, seo_score, category_id
FROM products
WHERE shop_id = $1
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY seo_score ASC
LIMIT $2 OFFSET $3;

-- name: CountProductsForSEOAudit :one
SELECT COUNT(*) FROM products
WHERE shop_id = $1
  AND status = 'active'
  AND deleted_at IS NULL;

-- name: GetAllProductsForSEOAudit :many
SELECT id, title, slug, seo_title, seo_description, description,
       review_count, price, seo_score, category_id
FROM products
WHERE shop_id = $1
  AND status = 'active'
  AND deleted_at IS NULL;

-- ── Sitemap ───────────────────────────────────────────────────────────────────

-- name: GetProductsForSitemap :many
SELECT id, slug, created_at, avg_rating, review_count
FROM products
WHERE shop_id = $1
  AND status = 'active'
  AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetCategoriesForSitemap :many
SELECT id, slug FROM categories
WHERE shop_id = $1
ORDER BY slug;

-- name: UpdateProductSEOFields :exec
UPDATE products
SET seo_title = $3, seo_description = $4
WHERE id = $1 AND shop_id = $2;
