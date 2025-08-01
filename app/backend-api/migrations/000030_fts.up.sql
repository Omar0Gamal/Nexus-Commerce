-- Extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- Full-text search column
ALTER TABLE products ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;

-- GIN indexes
CREATE INDEX IF NOT EXISTS idx_products_search_vector ON products USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_products_title_trgm    ON products USING GIN(title gin_trgm_ops);

-- Backfill existing rows
UPDATE products
SET search_vector = to_tsvector('english',
    COALESCE(title, '') || ' ' ||
    COALESCE(description, '') || ' ' ||
    COALESCE(sku, '')
);

-- Auto-update trigger
CREATE OR REPLACE FUNCTION products_search_vector_update()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    NEW.search_vector := to_tsvector('english',
        COALESCE(NEW.title, '') || ' ' ||
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(NEW.sku, '')
    );
    RETURN NEW;
END;
$$;

CREATE TRIGGER products_search_vector_trigger
    BEFORE INSERT OR UPDATE OF title, description, sku ON products
    FOR EACH ROW EXECUTE FUNCTION products_search_vector_update();

-- Dynamic facet attributes
CREATE TABLE IF NOT EXISTS product_attributes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id    UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    key        TEXT NOT NULL,
    value      TEXT NOT NULL,
    UNIQUE(product_id, key)
);
CREATE INDEX IF NOT EXISTS idx_product_attrs_facet ON product_attributes(shop_id, key, value);
