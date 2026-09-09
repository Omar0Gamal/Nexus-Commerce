-- Phase 17: Advanced Inventory Management
-- mirrors migrations/000023_inventory_advanced.up.sql for sqlc

CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    address     TEXT,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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

CREATE TABLE product_bundles (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id              UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    bundle_product_id    UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    component_variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity             SMALLINT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    UNIQUE (bundle_product_id, component_variant_id)
);

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
