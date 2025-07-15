DROP INDEX IF EXISTS idx_customers_active;
DROP INDEX IF EXISTS idx_products_active;

ALTER TABLE customers DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE products  DROP COLUMN IF EXISTS deleted_at;
