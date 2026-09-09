-- sql/schema/028_subscriptions.sql
-- Subscription plans, subscriptions and subscription orders for sqlc

CREATE TABLE subscription_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id         UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    product_id      UUID,
    price           NUMERIC(12,2) NOT NULL,
    billing_cycle   VARCHAR(20) NOT NULL DEFAULT 'monthly',
    trial_days      INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id               UUID NOT NULL,
    customer_id           UUID NOT NULL,
    plan_id               UUID NOT NULL,
    status                VARCHAR(20) NOT NULL DEFAULT 'active',
    current_period_start  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end    TIMESTAMPTZ NOT NULL,
    next_billing_at       TIMESTAMPTZ NOT NULL,
    trial_end             TIMESTAMPTZ,
    cancelled_at          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subscription_orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL,
    order_id        UUID,
    billing_date    TIMESTAMPTZ NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
