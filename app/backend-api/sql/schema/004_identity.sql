-- ==================== Identity & Security Layer ====================

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
