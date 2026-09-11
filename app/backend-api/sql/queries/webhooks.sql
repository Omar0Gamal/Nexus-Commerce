-- ==================== Webhooks ====================

-- name: CreateWebhook :one
INSERT INTO webhooks (
    shop_id, topic, target_url, secret_key, is_active
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetWebhook :one
SELECT * FROM webhooks
WHERE id = $1 AND shop_id = $2;

-- name: ListWebhooks :many
SELECT * FROM webhooks
WHERE shop_id = $1
ORDER BY topic ASC;

-- name: ListActiveWebhooks :many
SELECT * FROM webhooks
WHERE shop_id = $1 AND is_active = TRUE
ORDER BY topic ASC;

-- name: ListWebhooksByTopic :many
SELECT * FROM webhooks
WHERE shop_id = $1 AND topic = $2 AND is_active = TRUE;

-- name: UpdateWebhook :one
UPDATE webhooks SET
    topic = COALESCE(sqlc.narg('topic'), topic),
    target_url = COALESCE(sqlc.narg('target_url'), target_url),
    secret_key = COALESCE(sqlc.narg('secret_key'), secret_key),
    is_active = COALESCE(sqlc.narg('is_active'), is_active)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: ToggleWebhook :one
UPDATE webhooks SET is_active = NOT is_active
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: DeleteWebhook :exec
DELETE FROM webhooks WHERE id = $1 AND shop_id = $2;
