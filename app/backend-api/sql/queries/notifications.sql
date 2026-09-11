-- name: CreateNotification :one
INSERT INTO notifications (shop_id, recipient_type, recipient_id, channel, event_type, title, body, action_url, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetNotifications :many
SELECT * FROM notifications
WHERE shop_id = $1 AND recipient_type = $2 AND recipient_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetUnreadCount :one
SELECT COUNT(*) AS count FROM notifications
WHERE shop_id = $1 AND recipient_type = $2 AND recipient_id = $3 AND is_read = FALSE;

-- name: MarkNotificationRead :exec
UPDATE notifications SET is_read = TRUE, read_at = NOW()
WHERE id = $1 AND shop_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET is_read = TRUE, read_at = NOW()
WHERE shop_id = $1 AND recipient_type = $2 AND recipient_id = $3 AND is_read = FALSE;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE shop_id = $1 AND created_at < NOW() - INTERVAL '90 days';

-- name: UpsertPushSubscription :exec
INSERT INTO push_subscriptions (shop_id, user_type, user_id, endpoint, p256dh, auth_key, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (endpoint) DO UPDATE
SET p256dh = EXCLUDED.p256dh, auth_key = EXCLUDED.auth_key;

-- name: GetPushSubscriptions :many
SELECT * FROM push_subscriptions
WHERE shop_id = $1 AND user_type = $2 AND user_id = $3;

-- name: DeletePushSubscription :exec
DELETE FROM push_subscriptions WHERE endpoint = $1 AND shop_id = $2;

-- name: GetCustomerNotifPrefs :one
SELECT * FROM customer_notification_prefs WHERE customer_id = $1;

-- name: UpsertCustomerNotifPrefs :one
INSERT INTO customer_notification_prefs (customer_id, email_order_updates, email_marketing, push_order_updates, push_marketing, in_app_all)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (customer_id) DO UPDATE
SET email_order_updates = EXCLUDED.email_order_updates,
    email_marketing = EXCLUDED.email_marketing,
    push_order_updates = EXCLUDED.push_order_updates,
    push_marketing = EXCLUDED.push_marketing,
    in_app_all = EXCLUDED.in_app_all,
    updated_at = NOW()
RETURNING *;
