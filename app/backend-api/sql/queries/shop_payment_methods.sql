-- ==================== Shop Payment Methods ====================

-- name: CreateShopPaymentMethod :one
INSERT INTO shop_payment_methods (
    shop_id, provider, is_enabled, encrypted_credentials
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetShopPaymentMethod :one
SELECT * FROM shop_payment_methods
WHERE id = $1 AND shop_id = $2;

-- name: GetShopPaymentMethodByProvider :one
SELECT * FROM shop_payment_methods
WHERE shop_id = $1 AND provider = $2;

-- name: ListShopPaymentMethods :many
SELECT * FROM shop_payment_methods
WHERE shop_id = $1
ORDER BY provider ASC;

-- name: ListEnabledPaymentMethods :many
SELECT * FROM shop_payment_methods
WHERE shop_id = $1 AND is_enabled = TRUE
ORDER BY provider ASC;

-- name: UpdateShopPaymentMethod :one
UPDATE shop_payment_methods SET
    is_enabled = COALESCE(sqlc.narg('is_enabled'), is_enabled),
    encrypted_credentials = COALESCE(sqlc.narg('encrypted_credentials'), encrypted_credentials)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: TogglePaymentMethod :one
UPDATE shop_payment_methods SET is_enabled = NOT is_enabled
WHERE id = $1 AND shop_id = $2
RETURNING *;

-- name: DeleteShopPaymentMethod :exec
DELETE FROM shop_payment_methods WHERE id = $1 AND shop_id = $2;
