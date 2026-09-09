-- Phase 13 rollback

DROP TABLE IF EXISTS social_proof_events;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS wishlists;
ALTER TABLE products
    DROP COLUMN IF EXISTS review_count,
    DROP COLUMN IF EXISTS avg_rating;
