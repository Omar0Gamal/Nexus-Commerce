-- For sqlc schema: we represent the products table without the trigger
-- and with the search_vector column that will be controlled by the trigger.
-- The product_attributes table is separate.

ALTER TABLE products ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;

CREATE TABLE IF NOT EXISTS product_attributes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id    UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    key        TEXT NOT NULL,
    value      TEXT NOT NULL,
    UNIQUE(product_id, key)
);
