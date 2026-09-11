-- name: CreateKBArticle :one
INSERT INTO kb_articles (shop_id, category, title, slug, body, is_published, position)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetKBArticleBySlug :one
SELECT * FROM kb_articles
WHERE shop_id = $1 AND slug = $2 AND is_published = true;

-- name: GetKBArticleByID :one
SELECT * FROM kb_articles
WHERE id = $1 AND shop_id = $2;

-- name: ListKBArticles :many
SELECT * FROM kb_articles
WHERE shop_id = $1
  AND (sqlc.narg('category')::text IS NULL OR category = sqlc.narg('category'))
  AND (sqlc.narg('published_only')::boolean IS NULL OR is_published = sqlc.narg('published_only'))
ORDER BY position ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateKBArticle :one
UPDATE kb_articles SET
    category     = COALESCE(sqlc.narg('category'), category),
    title        = COALESCE(sqlc.narg('title'), title),
    slug         = COALESCE(sqlc.narg('slug'), slug),
    body         = COALESCE(sqlc.narg('body'), body),
    is_published = COALESCE(sqlc.narg('is_published'), is_published),
    position     = COALESCE(sqlc.narg('position'), position),
    updated_at   = NOW()
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: DeleteKBArticle :exec
DELETE FROM kb_articles WHERE id = $1 AND shop_id = $2;

-- name: IncrementKBViewCount :exec
UPDATE kb_articles SET view_count = view_count + 1 WHERE id = $1;

-- name: IncrementKBHelpfulCount :exec
UPDATE kb_articles SET helpful_count = helpful_count + 1 WHERE id = $1;
