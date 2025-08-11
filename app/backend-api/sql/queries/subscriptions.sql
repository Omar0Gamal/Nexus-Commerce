-- sql/queries/subscriptions.sql

-- name: CreateSubscriptionPlan :one
INSERT INTO subscription_plans (shop_id, name, description, product_id, price, billing_cycle, trial_days, is_active)
VALUES (@shop_id, @name, @description, @product_id, @price, @billing_cycle, @trial_days, @is_active)
RETURNING *;

-- name: GetSubscriptionPlan :one
SELECT * FROM subscription_plans
WHERE id = @id AND shop_id = @shop_id;

-- name: ListSubscriptionPlans :many
SELECT * FROM subscription_plans
WHERE shop_id = @shop_id AND is_active = TRUE
ORDER BY created_at DESC;

-- name: UpdateSubscriptionPlan :one
UPDATE subscription_plans
SET name = @name, description = @description, price = @price,
    billing_cycle = @billing_cycle, trial_days = @trial_days, is_active = @is_active,
    updated_at = NOW()
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: DeleteSubscriptionPlan :exec
UPDATE subscription_plans SET is_active = FALSE, updated_at = NOW()
WHERE id = @id AND shop_id = @shop_id;

-- name: CreateSubscription :one
INSERT INTO subscriptions (shop_id, customer_id, plan_id, status, current_period_start, current_period_end, next_billing_at, trial_end)
VALUES (@shop_id, @customer_id, @plan_id, @status, @current_period_start, @current_period_end, @next_billing_at, @trial_end)
RETURNING *;

-- name: GetSubscription :one
SELECT s.*, sp.name AS plan_name, sp.price AS plan_price, sp.billing_cycle
FROM subscriptions s
JOIN subscription_plans sp ON sp.id = s.plan_id
WHERE s.id = @id AND s.shop_id = @shop_id;

-- name: ListCustomerSubscriptions :many
SELECT s.*, sp.name AS plan_name, sp.price AS plan_price, sp.billing_cycle
FROM subscriptions s
JOIN subscription_plans sp ON sp.id = s.plan_id
WHERE s.customer_id = @customer_id AND s.shop_id = @shop_id
ORDER BY s.created_at DESC;

-- name: ListShopSubscriptions :many
SELECT s.*, sp.name AS plan_name, sp.price AS plan_price, sp.billing_cycle
FROM subscriptions s
JOIN subscription_plans sp ON sp.id = s.plan_id
WHERE s.shop_id = @shop_id
ORDER BY s.created_at DESC
LIMIT @lim OFFSET @off;

-- name: UpdateSubscriptionStatus :one
UPDATE subscriptions
SET status = @status, cancelled_at = @cancelled_at, updated_at = NOW()
WHERE id = @id AND shop_id = @shop_id
RETURNING *;

-- name: UpdateSubscriptionBilling :one
UPDATE subscriptions
SET next_billing_at = @next_billing_at,
    current_period_start = @current_period_start,
    current_period_end = @current_period_end,
    status = @status,
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: GetDueSubscriptions :many
SELECT s.*, sp.price AS plan_price, sp.billing_cycle, sp.product_id AS plan_product_id
FROM subscriptions s
JOIN subscription_plans sp ON sp.id = s.plan_id
WHERE s.next_billing_at <= NOW()
  AND s.status IN ('active', 'trialing')
ORDER BY s.next_billing_at
LIMIT 100;

-- name: CreateSubscriptionOrder :one
INSERT INTO subscription_orders (subscription_id, order_id, billing_date, status)
VALUES (@subscription_id, @order_id, @billing_date, @status)
RETURNING *;

-- name: UpdateSubscriptionOrderStatus :one
UPDATE subscription_orders
SET status = @status
WHERE id = @id
RETURNING *;

-- name: GetFailedSubscriptionOrders :many
SELECT so.*, s.customer_id, s.shop_id
FROM subscription_orders so
JOIN subscriptions s ON s.id = so.subscription_id
WHERE so.status = 'failed'
  AND so.billing_date > NOW() - INTERVAL '7 days'
ORDER BY so.billing_date DESC;
