-- ==================== Shipping Zones ====================

-- name: CreateShippingZone :one
INSERT INTO shipping_zones (
    shop_id, name, regions
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetShippingZone :one
SELECT * FROM shipping_zones
WHERE id = $1 AND shop_id = $2;

-- name: ListShippingZones :many
SELECT * FROM shipping_zones
WHERE shop_id = $1
ORDER BY name ASC;

-- name: UpdateShippingZone :one
UPDATE shipping_zones SET
    name = COALESCE(sqlc.narg('name'), name),
    regions = COALESCE(sqlc.narg('regions'), regions)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteShippingZone :exec
DELETE FROM shipping_zones WHERE id = $1 AND shop_id = $2;

-- ==================== Shipping Rates ====================

-- name: CreateShippingRate :one
INSERT INTO shipping_rates (zone_id, shop_id, name, rate_type, base_rate, rate_per_kg, min_order_free_shipping, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetShippingRate :one
SELECT * FROM shipping_rates
WHERE id = $1 AND shop_id = $2;

-- name: ListRatesByZone :many
SELECT * FROM shipping_rates
WHERE zone_id = $1 AND shop_id = $2
ORDER BY base_rate ASC;

-- name: ListShippingRates :many
SELECT * FROM shipping_rates
WHERE shop_id = $1
ORDER BY name ASC;

-- name: UpdateShippingRate :one
UPDATE shipping_rates
SET name                    = COALESCE(sqlc.narg('name'), name),
    rate_type               = COALESCE(sqlc.narg('rate_type'), rate_type),
    base_rate               = COALESCE(sqlc.narg('base_rate'), base_rate),
    rate_per_kg             = COALESCE(sqlc.narg('rate_per_kg'), rate_per_kg),
    min_order_free_shipping = COALESCE(sqlc.narg('min_order_free_shipping'), min_order_free_shipping),
    is_active               = COALESCE(sqlc.narg('is_active'), is_active)
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteShippingRate :exec
DELETE FROM shipping_rates WHERE id = $1 AND shop_id = $2;

-- name: FindRateByCountry :one
-- Find the cheapest active rate whose zone covers the given country code.
SELECT sr.*
FROM shipping_rates sr
JOIN shipping_zones sz ON sz.id = sr.zone_id
WHERE sz.shop_id = $1
  AND sr.is_active = TRUE
  AND (
      sz.regions = '[]'::jsonb
      OR sz.regions @> jsonb_build_array($2::text)
  )
ORDER BY sr.base_rate ASC
LIMIT 1;
