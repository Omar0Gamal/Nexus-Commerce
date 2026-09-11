-- name: CreateWarehouse :one
INSERT INTO warehouses (shop_id, name, address, is_default)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetWarehouses :many
SELECT * FROM warehouses WHERE shop_id = $1 AND is_active = TRUE ORDER BY is_default DESC, name;

-- name: GetInventoryLevel :one
SELECT * FROM inventory_levels
WHERE shop_id = $1 AND variant_id = $2 AND warehouse_id = $3;

-- name: GetInventoryLevelsByVariant :many
SELECT * FROM inventory_levels
WHERE shop_id = $1 AND variant_id = $2;

-- name: AdjustInventoryLevel :one
INSERT INTO inventory_levels (shop_id, variant_id, warehouse_id, quantity_on_hand, low_stock_threshold)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (variant_id, warehouse_id) DO UPDATE
SET quantity_on_hand = inventory_levels.quantity_on_hand + $4,
    updated_at = NOW()
RETURNING *;

-- name: ReserveStock :exec
UPDATE inventory_levels
SET quantity_reserved = quantity_reserved + $4,
    quantity_on_hand  = quantity_on_hand - $4,
    updated_at = NOW()
WHERE shop_id = $1 AND variant_id = $2 AND warehouse_id = $3
  AND quantity_on_hand >= $4;

-- name: ReleaseReservedStock :exec
UPDATE inventory_levels
SET quantity_reserved = GREATEST(0, quantity_reserved - $4),
    quantity_on_hand  = quantity_on_hand + $4,
    updated_at = NOW()
WHERE shop_id = $1 AND variant_id = $2 AND warehouse_id = $3;

-- name: RecordStockMovement :one
INSERT INTO stock_movements (shop_id, variant_id, warehouse_id, movement_type, quantity_delta, reference_id, reference_type, note, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: GetStockMovements :many
SELECT * FROM stock_movements
WHERE shop_id = $1 AND variant_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetLowStockVariants :many
SELECT il.variant_id, il.quantity_on_hand, il.low_stock_threshold, il.warehouse_id
FROM inventory_levels il
WHERE il.shop_id = $1 AND il.quantity_on_hand <= il.low_stock_threshold
ORDER BY il.quantity_on_hand ASC;

-- name: GetProductBundleComponents :many
SELECT * FROM product_bundles
WHERE shop_id = $1 AND bundle_product_id = $2;

-- name: UpsertBundleComponent :one
INSERT INTO product_bundles (shop_id, bundle_product_id, component_variant_id, quantity)
VALUES ($1, $2, $3, $4)
ON CONFLICT (bundle_product_id, component_variant_id) DO UPDATE
SET quantity = EXCLUDED.quantity
RETURNING *;

-- name: GetActiveReorderRules :many
SELECT * FROM reorder_rules
WHERE shop_id = $1 AND is_active = TRUE;

-- name: UpsertReorderRule :one
INSERT INTO reorder_rules (shop_id, variant_id, reorder_point, reorder_qty, supplier_email)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (shop_id, variant_id) DO UPDATE
SET reorder_point = EXCLUDED.reorder_point,
    reorder_qty   = EXCLUDED.reorder_qty,
    supplier_email = EXCLUDED.supplier_email
RETURNING *;

-- name: MarkReorderRuleTriggered :exec
UPDATE reorder_rules SET last_triggered_at = NOW()
WHERE id = $1;
