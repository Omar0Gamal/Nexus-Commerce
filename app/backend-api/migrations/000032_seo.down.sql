-- migrations/000032_seo.down.sql
ALTER TABLE categories
    DROP COLUMN IF EXISTS seo_description,
    DROP COLUMN IF EXISTS seo_title;

ALTER TABLE products
    DROP COLUMN IF EXISTS structured_data,
    DROP COLUMN IF EXISTS seo_score;

DROP TABLE IF EXISTS url_redirects;
