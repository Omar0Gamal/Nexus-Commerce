DROP TABLE IF EXISTS reorder_rules;
ALTER TABLE product_variants
    DROP COLUMN IF EXISTS allows_pre_order,
    DROP COLUMN IF EXISTS pre_order_message,
    DROP COLUMN IF EXISTS pre_order_ships_at;
DROP TABLE IF EXISTS product_bundles;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS inventory_levels;
DROP TABLE IF EXISTS warehouses;
