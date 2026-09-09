CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_products_title_trgm
    ON products USING GIN (title gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_products_description_trgm
    ON products USING GIN (description gin_trgm_ops)
    WHERE deleted_at IS NULL;
