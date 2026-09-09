-- ==================== Phase 13: Wishlists, Reviews & Social Proof ====================

CREATE TABLE wishlists (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id  UUID REFERENCES product_variants(id) ON DELETE SET NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (shop_id, customer_id, product_id)
);
CREATE INDEX idx_wishlists_shop_customer ON wishlists(shop_id, customer_id);
CREATE INDEX idx_wishlists_shop_product  ON wishlists(shop_id, product_id);

CREATE TABLE product_reviews (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id              UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    product_id           UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    customer_id          UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    order_id             UUID REFERENCES orders(id) ON DELETE SET NULL,
    rating               SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title                TEXT,
    body                 TEXT,
    status               TEXT NOT NULL DEFAULT 'pending'
                             CHECK (status IN ('pending', 'approved', 'rejected')),
    is_verified_purchase BOOLEAN NOT NULL DEFAULT FALSE,
    helpful_count        INTEGER NOT NULL DEFAULT 0,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (shop_id, product_id, customer_id)
);
CREATE INDEX idx_reviews_shop_product_status ON product_reviews(shop_id, product_id, status);
CREATE INDEX idx_reviews_shop_status         ON product_reviews(shop_id, status);

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS review_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS avg_rating   NUMERIC(3,2) NOT NULL DEFAULT 0;

CREATE TABLE social_proof_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id    UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    city       TEXT,
    country    TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_social_proof_shop ON social_proof_events(shop_id, created_at DESC);
