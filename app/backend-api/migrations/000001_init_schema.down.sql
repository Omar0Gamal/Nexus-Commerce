-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS shop_payment_methods;
DROP TABLE IF EXISTS shipping_zones;
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS support_automation_rules;
DROP TABLE IF EXISTS support_messages;
DROP TABLE IF EXISTS support_tickets;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS shop_staff;
DROP TABLE IF EXISTS shop_roles;
DROP TABLE IF EXISTS shop_invoices;
DROP TABLE IF EXISTS shops;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS user_secrets;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS platform_admins;
DROP TABLE IF EXISTS plans;

-- Drop indexes (most are dropped with tables, but explicit for safety)
DROP INDEX IF EXISTS idx_audit_logs_shop_created;
DROP INDEX IF EXISTS idx_tickets_assigned_staff;
DROP INDEX IF EXISTS idx_tickets_shop_status;
DROP INDEX IF EXISTS idx_orders_shop_customer;
DROP INDEX IF EXISTS idx_shops_subdomain;

-- Drop enum types
DROP TYPE IF EXISTS message_sender;
DROP TYPE IF EXISTS ticket_priority;
DROP TYPE IF EXISTS ticket_status;
DROP TYPE IF EXISTS invoice_status;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS payment_provider;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS invite_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS shop_status;
