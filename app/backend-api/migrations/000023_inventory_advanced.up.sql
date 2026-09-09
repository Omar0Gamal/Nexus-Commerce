-- Phase 17: Advanced Inventory Management

-- Warehouses / stock locations
CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    address     TEXT,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_warehouses_shop ON warehouses(shop_id);

-- Per-warehouse stock levels
CREATE TABLE inventory_levels (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id             UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    variant_id          UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    warehouse_id        UUID NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    quantity_on_hand    INT NOT NULL DEFAULT 0 CHECK (quantity_on_hand >= 0),
    quantity_reserved   INT NOT NULL DEFAULT 0 CHECK (quantity_reserved >= 0),
    quantity_incoming   INT NOT NULL DEFAULT 0 CHECK (quantity_incoming >= 0),
    low_stock_threshold INT NOT NULL DEFAULT 5,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (variant_id, warehouse_id)
);
CREATE INDEX idx_inv_levels_shop_variant ON inventory_levels(shop_id, variant_id);

-- Stock movement audit log
CREATE TABLE stock_movements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    warehouse_id    UUID REFERENCES warehouses(id),
    movement_type   TEXT NOT NULL CHECK (movement_type IN ('sale','restock','adjustment','transfer','return','pre_order_reserve')),
    quantity_delta  INT NOT NULL,
    reference_id    UUID,
    reference_type  TEXT,
    note            TEXT,
    created_by      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_stock_movements_variant   ON stock_movements(shop_id, variant_id, created_at DESC);
CREATE INDEX idx_stock_movements_warehouse ON stock_movements(shop_id, warehouse_id, created_at DESC);

-- Bundle / kit products
CREATE TABLE product_bundles (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id              UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    bundle_product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    component_variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity             SMALLINT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    UNIQUE (bundle_product_id, component_variant_id)
);

-- Pre-order support
ALTER TABLE product_variants
    ADD COLUMN IF NOT EXISTS allows_pre_order  BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS pre_order_message TEXT,
    ADD COLUMN IF NOT EXISTS pre_order_ships_at DATE;

-- Reorder automation rules
CREATE TABLE reorder_rules (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id           UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    variant_id        UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    reorder_point     INT NOT NULL DEFAULT 10,
    reorder_qty       INT NOT NULL DEFAULT 50,
    supplier_email    TEXT,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    UNIQUE (shop_id, variant_id)
);
