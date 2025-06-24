-- Reverse migration 003

ALTER TABLE order_items DROP COLUMN IF EXISTS variant_id;

DROP INDEX IF EXISTS idx_images_variant;
DROP INDEX IF EXISTS idx_images_product;
DROP TABLE IF EXISTS product_images;

DROP INDEX IF EXISTS idx_variants_product;
DROP TABLE IF EXISTS product_variants;

ALTER TABLE products DROP COLUMN IF EXISTS description;
