-- ==================== Tenant Layer (Shops) ====================

CREATE TABLE shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID REFERENCES plans(id) NOT NULL,
    owner_user_id UUID NOT NULL,

    -- Identity
    name VARCHAR(100) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    custom_domain VARCHAR(255) UNIQUE,

    -- Configuration
    status shop_status DEFAULT 'active',
    currency VARCHAR(3) DEFAULT 'EGP',
    timezone VARCHAR(50) DEFAULT 'Africa/Cairo',

    -- Billing Status
    current_period_end TIMESTAMP WITH TIME ZONE,
    is_overdue BOOLEAN DEFAULT FALSE,

    -- SEO
    seo_title VARCHAR(255),
    seo_description TEXT,
    favicon_url TEXT,

    -- Notification preferences
    notification_prefs JSONB DEFAULT '{"new_orders": true, "low_stock": true, "customer_reviews": true}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_shops_subdomain ON shops(subdomain);

CREATE TABLE shop_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,

    amount DECIMAL(10, 2) NOT NULL,
    status invoice_status NOT NULL,
    billing_reason VARCHAR(50),

    hosted_invoice_url VARCHAR(512),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
