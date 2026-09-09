DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;
DROP FUNCTION IF EXISTS products_search_vector_update();
DROP INDEX IF EXISTS idx_products_search_vector;
DROP INDEX IF EXISTS idx_products_title_trgm;
ALTER TABLE products DROP COLUMN IF EXISTS search_vector;
DROP TABLE IF EXISTS product_attributes;
