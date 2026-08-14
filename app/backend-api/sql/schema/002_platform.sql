-- ==================== Platform Layer ====================

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    monthly_price DECIMAL(10, 2) NOT NULL,

    -- Limits & Quotas
    max_products INTEGER DEFAULT 100,
    max_staff_accounts INTEGER DEFAULT 1,
    max_storage_mb INTEGER DEFAULT 512,
    transaction_fee_percent DECIMAL(5, 2) DEFAULT 2.0,

    -- Feature Flags
    features JSONB DEFAULT '{"custom_domain": false, "support_automation": false}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE platform_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    is_2fa_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
