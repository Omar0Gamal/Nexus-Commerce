-- sql/schema/029_seo.sql
-- Phase 31: SEO Foundation — for sqlc code generation

CREATE TABLE url_redirects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL,
    from_path   TEXT NOT NULL,
    to_path     TEXT NOT NULL,
    status_code SMALLINT NOT NULL DEFAULT 301,
    hit_count   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, from_path)
);

-- Note: seo_score, structured_data, categories.seo_title/seo_description are
-- added via ALTER TABLE in the migration. sqlc reads them from the full schema
-- composed across all schema files; separate ALTER stubs are not needed here
-- because products and categories are already declared in earlier schema files.
