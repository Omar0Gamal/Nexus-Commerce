-- =============================================
-- Migration 003: Product Variants, Product Images, Description
-- Adds variant support (size/color combos), image uploads,
-- and links order_items to specific variants.
-- =============================================

-- Add description to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS description TEXT;

-- ── Product Variants ────────────────────────────────────────────────────────
-- Each variant represents a purchasable combination (e.g. "Red / Large").
-- `options` stores the option key/value pairs as JSONB: {"Color":"Red","Size":"L"}

CREATE TABLE product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,

    title VARCHAR(255) NOT NULL,         -- e.g. "Red / Large"
    sku VARCHAR(100),
    price DECIMAL(10, 2) NOT NULL,
    compare_at_price DECIMAL(10, 2),
    stock_quantity INTEGER NOT NULL DEFAULT 0,

    options JSONB DEFAULT '{}',          -- {"Color":"Red","Size":"L"}
    position INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(shop_id, sku)
);

CREATE INDEX idx_variants_product ON product_variants(product_id);

-- ── Product Images ──────────────────────────────────────────────────────────
-- Images can be product-level (variant_id IS NULL) or variant-specific.

CREATE TABLE product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    shop_id UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    variant_id UUID REFERENCES product_variants(id) ON DELETE SET NULL,

    url VARCHAR(512) NOT NULL,           -- e.g. "/uploads/{shop_id}/{product_id}/{filename}"
    alt_text VARCHAR(255),
    position INTEGER DEFAULT 0,
    is_primary BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_images_product ON product_images(product_id);
CREATE INDEX idx_images_variant ON product_images(variant_id);

-- ── Link order items to variants ────────────────────────────────────────────
ALTER TABLE order_items ADD COLUMN IF NOT EXISTS variant_id UUID REFERENCES product_variants(id) ON DELETE SET NULL;
