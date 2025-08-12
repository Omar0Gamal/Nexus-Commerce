-- migrations/000032_seo.up.sql
-- Phase 31: SEO Foundation — URL redirects, SEO scoring, structured data

-- URL redirect manager (e.g. after slug changes)
CREATE TABLE url_redirects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    from_path   TEXT NOT NULL,
    to_path     TEXT NOT NULL,
    status_code SMALLINT NOT NULL DEFAULT 301 CHECK (status_code IN (301, 302)),
    hit_count   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, from_path)
);

CREATE INDEX idx_redirects_shop ON url_redirects(shop_id, from_path);

-- SEO score cache on products (0-100, recomputed by nightly worker)
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS seo_score SMALLINT NOT NULL DEFAULT 0;

-- Structured data override per product (optional JSON-LD blob)
ALTER TABLE products
    ADD COLUMN IF NOT EXISTS structured_data JSONB;

-- SEO metadata on categories
ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS seo_title       TEXT,
    ADD COLUMN IF NOT EXISTS seo_description TEXT;
