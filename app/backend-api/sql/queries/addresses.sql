-- ==================== Customer Addresses ====================

-- name: ListCustomerAddresses :many
SELECT * FROM customer_addresses
WHERE customer_id = @customer_id AND shop_id = @shop_id
ORDER BY is_default DESC, created_at ASC;

-- name: GetCustomerAddress :one
SELECT * FROM customer_addresses
WHERE id = @id AND customer_id = @customer_id AND shop_id = @shop_id;

-- name: CountCustomerAddresses :one
SELECT COUNT(*) FROM customer_addresses
WHERE customer_id = @customer_id AND shop_id = @shop_id;

-- name: CreateCustomerAddress :one
INSERT INTO customer_addresses (
    customer_id, shop_id, label, first_name, last_name,
    address1, address2, city, state, country, zip_code, phone, is_default
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: UpdateCustomerAddress :one
UPDATE customer_addresses
SET
    label      = COALESCE(sqlc.narg('label'), label),
    first_name = COALESCE(sqlc.narg('first_name'), first_name),
    last_name  = COALESCE(sqlc.narg('last_name'), last_name),
    address1   = COALESCE(sqlc.narg('address1'), address1),
    address2   = COALESCE(sqlc.narg('address2'), address2),
    city       = COALESCE(sqlc.narg('city'), city),
    state      = COALESCE(sqlc.narg('state'), state),
    country    = COALESCE(sqlc.narg('country'), country),
    zip_code   = COALESCE(sqlc.narg('zip_code'), zip_code),
    phone      = COALESCE(sqlc.narg('phone'), phone)
WHERE id = @id AND customer_id = @customer_id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteCustomerAddress :exec
DELETE FROM customer_addresses
WHERE id = @id AND customer_id = @customer_id AND shop_id = @shop_id;

-- name: ClearDefaultAddresses :exec
UPDATE customer_addresses
SET is_default = FALSE
WHERE customer_id = @customer_id AND shop_id = @shop_id;

-- name: SetAddressAsDefault :exec
UPDATE customer_addresses
SET is_default = TRUE
WHERE id = @id AND customer_id = @customer_id AND shop_id = @shop_id;
