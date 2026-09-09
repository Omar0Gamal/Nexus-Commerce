ALTER TABLE shops ADD COLUMN IF NOT EXISTS notification_prefs JSONB DEFAULT '{"new_orders": true, "low_stock": true, "customer_reviews": true}';
