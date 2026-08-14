-- Phase 15: Notifications System
-- mirrors migrations/000021_notifications.up.sql for sqlc

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    recipient_type  TEXT NOT NULL CHECK (recipient_type IN ('staff','customer')),
    recipient_id    UUID NOT NULL,
    channel         TEXT NOT NULL CHECK (channel IN ('in_app','email','push')),
    event_type      TEXT NOT NULL,
    title           TEXT NOT NULL,
    body            TEXT NOT NULL,
    action_url      TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    is_read         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at         TIMESTAMPTZ
);

CREATE TABLE push_subscriptions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    user_type   TEXT NOT NULL CHECK (user_type IN ('staff','customer')),
    user_id     UUID NOT NULL,
    endpoint    TEXT NOT NULL UNIQUE,
    p256dh      TEXT NOT NULL,
    auth_key    TEXT NOT NULL,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE customer_notification_prefs (
    customer_id         UUID PRIMARY KEY REFERENCES customers(id) ON DELETE CASCADE,
    email_order_updates BOOLEAN NOT NULL DEFAULT TRUE,
    email_marketing     BOOLEAN NOT NULL DEFAULT FALSE,
    push_order_updates  BOOLEAN NOT NULL DEFAULT TRUE,
    push_marketing      BOOLEAN NOT NULL DEFAULT FALSE,
    in_app_all          BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
