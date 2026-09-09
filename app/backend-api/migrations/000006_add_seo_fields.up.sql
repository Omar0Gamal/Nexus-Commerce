-- Add SEO meta fields to products, categories, and shops

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS seo_title VARCHAR(255),
    ADD COLUMN IF NOT EXISTS seo_description TEXT;

ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS meta_description TEXT;

ALTER TABLE shops
    ADD COLUMN IF NOT EXISTS seo_title VARCHAR(255),
    ADD COLUMN IF NOT EXISTS seo_description TEXT,
    ADD COLUMN IF NOT EXISTS favicon_url TEXT;
