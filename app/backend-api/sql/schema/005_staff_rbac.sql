-- ==================== Staff & RBAC Layer ====================

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
