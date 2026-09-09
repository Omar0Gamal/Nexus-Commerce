-- ==================== Order Returns ====================

CREATE TYPE return_status AS ENUM ('requested', 'approved', 'rejected', 'refunded');

CREATE TABLE order_returns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    order_id        UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    customer_id     UUID REFERENCES customers(id) ON DELETE SET NULL,

    reason          TEXT NOT NULL,
    status          return_status NOT NULL DEFAULT 'requested',

    paymob_refund_id    BIGINT,
    refund_amount       DECIMAL(12, 2),

    notes           TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_order_returns_shop     ON order_returns(shop_id);
CREATE INDEX idx_order_returns_order    ON order_returns(order_id);
CREATE INDEX idx_order_returns_customer ON order_returns(customer_id);
