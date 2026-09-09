-- =============================================
-- Migration 001: Extensions & Enum Types
-- =============================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Status Enums
CREATE TYPE shop_status AS ENUM ('active', 'suspended', 'maintenance', 'closed');
CREATE TYPE user_status AS ENUM ('active', 'banned', 'pending_invite');
CREATE TYPE invite_type AS ENUM ('new_store_owner', 'shop_staff');

-- Commerce Enums
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'processing', 'shipped', 'completed', 'cancelled', 'refunded');
CREATE TYPE payment_provider AS ENUM ('paymob', 'fawry', 'stripe', 'cash_on_delivery', 'vodafone_cash');
CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount', 'shipping_override');
CREATE TYPE invoice_status AS ENUM ('paid', 'open', 'void', 'uncollectible');

-- Support Enums
CREATE TYPE ticket_status AS ENUM ('open', 'pending_staff', 'pending_customer', 'resolved', 'closed');
CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high', 'urgent');
CREATE TYPE message_sender AS ENUM ('customer', 'staff', 'system');

-- =============================================
-- Platform Layer
-- =============================================

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    monthly_price DECIMAL(10, 2) NOT NULL,
    max_products INTEGER DEFAULT 100,
    max_staff_accounts INTEGER DEFAULT 1,
    max_storage_mb INTEGER DEFAULT 512,
    transaction_fee_percent DECIMAL(5, 2) DEFAULT 2.0,
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

-- =============================================
-- Identity & Security Layer
-- =============================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(100),
    phone VARCHAR(20),
    is_email_verified BOOLEAN DEFAULT FALSE,
    status user_status DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE user_secrets (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    totp_secret VARCHAR(255),
    backup_codes JSONB DEFAULT '[]',
    is_2fa_enabled BOOLEAN DEFAULT FALSE,
    recovery_email VARCHAR(255),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    token VARCHAR(64) UNIQUE NOT NULL,
    type invite_type NOT NULL,
    payload JSONB NOT NULL,
    invited_by_user_id UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================
-- Tenant Layer (Shops)
-- =============================================

CREATE TABLE shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID REFERENCES plans(id) NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    subdomain VARCHAR(63) UNIQUE NOT NULL,
    custom_domain VARCHAR(255) UNIQUE,
    status shop_status DEFAULT 'active',
    currency VARCHAR(3) DEFAULT 'EGP',
    timezone VARCHAR(50) DEFAULT 'Africa/Cairo',
    current_period_end TIMESTAMP WITH TIME ZONE,
    is_overdue BOOLEAN DEFAULT FALSE,
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

-- =============================================
-- Staff & RBAC Layer
-- =============================================

CREATE TABLE shop_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    parent_role_id UUID REFERENCES shop_roles(id),
    name VARCHAR(50) NOT NULL,
    permissions JSONB DEFAULT '[]',
    is_system_role BOOLEAN DEFAULT FALSE,
    UNIQUE(shop_id, name)
);

CREATE TABLE shop_staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES shop_roles(id),
    is_owner BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(shop_id, user_id)
);

-- =============================================
-- E-Commerce Core
-- =============================================

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    UNIQUE(shop_id, slug)
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    compare_at_price DECIMAL(10, 2),
    track_inventory BOOLEAN DEFAULT TRUE,
    sku VARCHAR(100),
    status VARCHAR(20) DEFAULT 'draft',
    UNIQUE(shop_id, slug),
    UNIQUE(shop_id, sku)
);

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(shop_id, email)
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) NOT NULL,
    customer_id UUID REFERENCES customers(id),
    order_number INTEGER NOT NULL,
    total_price DECIMAL(12, 2) NOT NULL,
    status order_status DEFAULT 'pending',
    shipping_address JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_shop_customer ON orders(shop_id, customer_id);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    quantity INTEGER NOT NULL
);

-- =============================================
-- Customer Support System
-- =============================================

CREATE TABLE support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    guest_email VARCHAR(255),
    subject VARCHAR(255) NOT NULL,
    status ticket_status DEFAULT 'open',
    priority ticket_priority DEFAULT 'medium',
    category VARCHAR(50),
    assigned_staff_id UUID REFERENCES shop_staff(id) ON DELETE SET NULL,
    related_order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tickets_shop_status ON support_tickets(shop_id, status);
CREATE INDEX idx_tickets_assigned_staff ON support_tickets(assigned_staff_id);

CREATE TABLE support_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID REFERENCES support_tickets(id) ON DELETE CASCADE,
    sender_type message_sender NOT NULL,
    staff_id UUID REFERENCES shop_staff(id) ON DELETE SET NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    message_body TEXT NOT NULL,
    attachments JSONB DEFAULT '[]',
    is_internal_note BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE support_automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    conditions JSONB NOT NULL,
    action_assign_staff_id UUID REFERENCES shop_staff(id),
    action_assign_priority ticket_priority,
    action_auto_reply_text TEXT,
    priority_order INTEGER DEFAULT 0
);

-- =============================================
-- Marketing & Logistics
-- =============================================

CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    type discount_type NOT NULL,
    value DECIMAL(10, 2) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE(shop_id, code)
);

CREATE TABLE shipping_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    regions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================
-- Integrations & Config
-- =============================================

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

-- =============================================
-- Audit & Security Logs
-- =============================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID REFERENCES shops(id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_name VARCHAR(255),
    ip_address INET,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    changes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_shop_created ON audit_logs(shop_id, created_at DESC);
