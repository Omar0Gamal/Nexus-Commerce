-- ==================== Integrations & Config ====================

CREATE TABLE shop_payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    provider payment_provider NOT NULL,
    is_enabled BOOLEAN DEFAULT FALSE,
    encrypted_credentials TEXT NOT NULL,
    UNIQUE(shop_id, provider)
);

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    topic VARCHAR(50) NOT NULL,
    target_url VARCHAR(512) NOT NULL,
    secret_key VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);
