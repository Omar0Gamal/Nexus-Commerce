ALTER TABLE products DROP COLUMN IF EXISTS seo_title, DROP COLUMN IF EXISTS seo_description;
ALTER TABLE categories DROP COLUMN IF EXISTS meta_description;
ALTER TABLE shops DROP COLUMN IF EXISTS seo_title, DROP COLUMN IF EXISTS seo_description, DROP COLUMN IF EXISTS favicon_url;
