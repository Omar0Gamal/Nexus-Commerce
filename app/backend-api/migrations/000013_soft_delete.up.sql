-- Add soft delete support to products and customers.
-- deleted_at IS NULL → active record
-- deleted_at IS NOT NULL → archived / soft-deleted

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Partial indexes to keep live-record lookups fast.
CREATE INDEX IF NOT EXISTS idx_products_active
    ON products (shop_id, status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_customers_active
    ON customers (shop_id)
    WHERE deleted_at IS NULL;
