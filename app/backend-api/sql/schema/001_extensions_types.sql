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
