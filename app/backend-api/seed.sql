-- ============================================================
-- Seed Data for Testing the Catalog Module
-- Run AFTER migrations: psql -U user -d saas_db -f seed.sql
-- ============================================================

-- 1. Create the 3 paid plans
INSERT INTO plans (id, name, monthly_price, max_products, max_staff_accounts, max_storage_mb, transaction_fee_percent, features)
VALUES
    (
        'a0000001-0000-0000-0000-000000000001',
        'basic',
        400.00,
        100,
        0,
        1024,
        1.5,
        '{"custom_domain":false,"coupons":true,"shipping_zones":true,"core_analytics":true,"wishlists":true,"product_reviews":true,"social_proof":false,"abandoned_cart_recovery":false,"notifications_in_app":true,"notifications_push":false,"notifications_sse":false,"multi_currency":false,"localization":true,"multi_warehouse":false,"inventory_bundles":false,"pre_orders":false,"reorder_automation":false,"flash_sales":false,"tiered_discounts":false,"loyalty_program":false,"referral_program":false,"pwa":true,"seo_basic":true,"seo_scoring":false,"seo_redirects":false,"seo_advanced":false,"webhooks":false,"advanced_analytics":false,"analytics_export":false,"realtime_analytics":false,"support_automation":false,"priority_support":false,"ai_text":false,"ai_seo":false,"ai_translate":false,"ai_email":false,"ai_reports":false,"ai_engine":false,"ai_vision":false}'
    ),
    (
        'a0000002-0000-0000-0000-000000000002',
        'professional',
        800.00,
        500,
        2,
        5120,
        1.0,
        '{"custom_domain":true,"coupons":true,"shipping_zones":true,"core_analytics":true,"realtime_analytics":true,"standard_analytics":true,"analytics_export":true,"wishlists":true,"product_reviews":true,"review_moderation":true,"social_proof":true,"abandoned_cart_recovery":true,"notifications_in_app":true,"notifications_push":true,"notifications_sse":true,"multi_currency":true,"localization":true,"multi_warehouse":false,"inventory_bundles":true,"pre_orders":true,"reorder_automation":false,"flash_sales":true,"tiered_discounts":true,"loyalty_program":false,"referral_program":false,"pwa":true,"seo_basic":true,"seo_scoring":true,"seo_redirects":true,"seo_advanced":false,"webhooks":true,"advanced_analytics":false,"support_automation":false,"priority_support":false,"ai_text":true,"ai_seo":true,"ai_translate":true,"ai_email":true,"ai_reports":false,"ai_engine":true,"ai_vision":false}'
    ),
    (
        'a0000003-0000-0000-0000-000000000003',
        'premium',
        1500.00,
        -1,
        10,
        20480,
        1.0,
        '{"custom_domain":true,"coupons":true,"shipping_zones":true,"core_analytics":true,"realtime_analytics":true,"standard_analytics":true,"analytics_export":true,"advanced_analytics":true,"analytics_alerts":true,"analytics_api_access":true,"wishlists":true,"product_reviews":true,"review_moderation":true,"social_proof":true,"abandoned_cart_recovery":true,"notifications_in_app":true,"notifications_push":true,"notifications_sse":true,"multi_currency":true,"localization":true,"multi_warehouse":true,"inventory_bundles":true,"pre_orders":true,"reorder_automation":true,"flash_sales":true,"tiered_discounts":true,"loyalty_program":true,"referral_program":true,"pwa":true,"seo_basic":true,"seo_scoring":true,"seo_redirects":true,"seo_advanced":true,"seo_search_console":true,"webhooks":true,"support_automation":true,"priority_support":true,"ai_text":true,"ai_seo":true,"ai_translate":true,"ai_email":true,"ai_reports":true,"ai_engine":true,"ai_vision":true,"ai_forecasting":true,"ai_churn_scoring":true}'
    )
ON CONFLICT (name) DO UPDATE SET
    monthly_price = EXCLUDED.monthly_price,
    max_products = EXCLUDED.max_products,
    max_staff_accounts = EXCLUDED.max_staff_accounts,
    max_storage_mb = EXCLUDED.max_storage_mb,
    transaction_fee_percent = EXCLUDED.transaction_fee_percent,
    features = EXCLUDED.features;

-- 2. Create test users (shop owners) — seed password hashes are intentionally opaque
INSERT INTO users (id, email, password_hash, full_name, is_email_verified, status)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'owner@coolshoes.com',
    '$2a$10$XRYAXLmRyHeAe1Z8ZiVH/OYMiy37EIx4flSha.m.sf7ryZ7WfFsau',
    'Cool Shoes Owner',
    true,
    'active'
) ON CONFLICT (id) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    is_email_verified = EXCLUDED.is_email_verified,
    status = EXCLUDED.status;

INSERT INTO users (id, email, password_hash, full_name, is_email_verified, status)
VALUES (
    'b0000000-0000-0000-0000-000000000002',
    'owner@techstore.com',
    '$2a$10$XRYAXLmRyHeAe1Z8ZiVH/OYMiy37EIx4flSha.m.sf7ryZ7WfFsau',
    'Tech Store Owner',
    true,
    'active'
) ON CONFLICT (id) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    is_email_verified = EXCLUDED.is_email_verified,
    status = EXCLUDED.status;

-- 3. Create Shop A: "Cool Shoes" (subdomain: coolshoes)
INSERT INTO shops (id, plan_id, owner_user_id, name, subdomain, custom_domain, status, currency, timezone)
VALUES (
    '11111111-aaaa-bbbb-cccc-000000000001',
    'a0000002-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000001',
    'Cool Shoes',
    'coolshoes',
    NULL,
    'active',
    'EGP',
    'Africa/Cairo'
) ON CONFLICT DO NOTHING;

-- 4. Create Shop B: "Tech Store" (subdomain: techstore)
INSERT INTO shops (id, plan_id, owner_user_id, name, subdomain, custom_domain, status, currency, timezone)
VALUES (
    '22222222-aaaa-bbbb-cccc-000000000002',
    'a0000003-0000-0000-0000-000000000003',
    'b0000000-0000-0000-0000-000000000002',
    'Tech Store',
    'techstore',
    NULL,
    'active',
    'EGP',
    'Africa/Cairo'
) ON CONFLICT DO NOTHING;

-- 5. Create categories for Shop A
INSERT INTO categories (id, shop_id, name, slug) VALUES
    ('cccc0001-0000-0000-0000-000000000001', '11111111-aaaa-bbbb-cccc-000000000001', 'Sneakers', 'sneakers'),
    ('cccc0001-0000-0000-0000-000000000002', '11111111-aaaa-bbbb-cccc-000000000001', 'Boots', 'boots')
ON CONFLICT DO NOTHING;

-- 6. Create categories for Shop B
INSERT INTO categories (id, shop_id, name, slug) VALUES
    ('cccc0002-0000-0000-0000-000000000001', '22222222-aaaa-bbbb-cccc-000000000002', 'Laptops', 'laptops'),
    ('cccc0002-0000-0000-0000-000000000002', '22222222-aaaa-bbbb-cccc-000000000002', 'Phones', 'phones')
ON CONFLICT DO NOTHING;

-- 7. Create products for Shop A (Cool Shoes)
INSERT INTO products (id, shop_id, category_id, title, slug, price, compare_at_price, sku, status, stock_quantity, track_inventory) VALUES
    ('dd000001-0000-0000-0000-000000000001', '11111111-aaaa-bbbb-cccc-000000000001', 'cccc0001-0000-0000-0000-000000000001',
     'Red Sneakers', 'red-sneakers', 199.99, 249.99, 'CS-RED-001', 'active', 100, true),
    ('dd000001-0000-0000-0000-000000000002', '11111111-aaaa-bbbb-cccc-000000000001', 'cccc0001-0000-0000-0000-000000000001',
     'Blue Sneakers', 'blue-sneakers', 179.99, NULL, 'CS-BLU-001', 'active', 75, true),
    ('dd000001-0000-0000-0000-000000000003', '11111111-aaaa-bbbb-cccc-000000000001', 'cccc0001-0000-0000-0000-000000000002',
     'Leather Boots', 'leather-boots', 349.99, 399.99, 'CS-LB-001', 'active', 50, true),
    ('dd000001-0000-0000-0000-000000000004', '11111111-aaaa-bbbb-cccc-000000000001', NULL,
     'Draft Sandals', 'draft-sandals', 89.99, NULL, NULL, 'draft', 0, false)
ON CONFLICT DO NOTHING;

-- 8. Create products for Shop B (Tech Store) — note: NO "red-sneakers" here!
INSERT INTO products (id, shop_id, category_id, title, slug, price, sku, status, stock_quantity, track_inventory) VALUES
    ('dd000002-0000-0000-0000-000000000001', '22222222-aaaa-bbbb-cccc-000000000002', 'cccc0002-0000-0000-0000-000000000001',
     'Gaming Laptop', 'gaming-laptop', 24999.99, 'TS-GL-001', 'active', 20, true),
    ('dd000002-0000-0000-0000-000000000002', '22222222-aaaa-bbbb-cccc-000000000002', 'cccc0002-0000-0000-0000-000000000002',
     'Smartphone X', 'smartphone-x', 14999.99, 'TS-SP-001', 'active', 30, true)
ON CONFLICT DO NOTHING;

-- 9. Create owner roles for each shop
INSERT INTO shop_roles (id, shop_id, name, permissions, is_system_role) VALUES
    ('ee000001-0000-0000-0000-000000000001', '11111111-aaaa-bbbb-cccc-000000000001', 'owner', '["*"]', true),
    ('ee000001-0000-0000-0000-000000000002', '22222222-aaaa-bbbb-cccc-000000000002', 'owner', '["*"]', true)
ON CONFLICT DO NOTHING;

-- 10. Add shop_staff entries (owners linked to roles)
INSERT INTO shop_staff (shop_id, user_id, role_id, is_owner) VALUES
    ('11111111-aaaa-bbbb-cccc-000000000001', 'b0000000-0000-0000-0000-000000000001', 'ee000001-0000-0000-0000-000000000001', true),
    ('22222222-aaaa-bbbb-cccc-000000000002', 'b0000000-0000-0000-0000-000000000002', 'ee000001-0000-0000-0000-000000000002', true)
ON CONFLICT DO NOTHING;

-- 11. Create a platform admin
INSERT INTO platform_admins (id, email, password_hash, full_name)
VALUES (
    'aa000000-0000-0000-0000-000000000001',
    'admin@nexus-commerce.com',
    '$2a$10$XRYAXLmRyHeAe1Z8ZiVH/OYMiy37EIx4flSha.m.sf7ryZ7WfFsau',
    'Platform Admin'
) ON CONFLICT DO NOTHING;

-- 12. Create customers for Shop A (Cool Shoes)
INSERT INTO customers (id, shop_id, email, password_hash, first_name, last_name, phone) VALUES
    ('ff000001-0000-0000-0000-000000000001', '11111111-aaaa-bbbb-cccc-000000000001',
    'alice@example.com', '$2a$10$XRYAXLmRyHeAe1Z8ZiVH/OYMiy37EIx4flSha.m.sf7ryZ7WfFsau',
     'Alice', 'Smith', '+201234567890'),
    ('ff000001-0000-0000-0000-000000000002', '11111111-aaaa-bbbb-cccc-000000000001',
    'bob@example.com', '$2a$10$XRYAXLmRyHeAe1Z8ZiVH/OYMiy37EIx4flSha.m.sf7ryZ7WfFsau',
     'Bob', 'Jones', NULL)
ON CONFLICT DO NOTHING;

-- 13. Create a customer for Shop B (Tech Store)
INSERT INTO customers (id, shop_id, email, password_hash, first_name, last_name) VALUES
    ('ff000002-0000-0000-0000-000000000001', '22222222-aaaa-bbbb-cccc-000000000002',
    'alice@example.com', '$2a$10$XRYAXLmRyHeAe1Z8ZiVH/OYMiy37EIx4flSha.m.sf7ryZ7WfFsau',
     'Alice', 'Smith')
ON CONFLICT DO NOTHING;

-- ============================================================
-- Summary:
--   Shop A (coolshoes): 2 categories, 4 products (3 active, 1 draft)
--   Shop B (techstore): 2 categories, 2 products (both active)
--   Platform Admin: admin@nexus-commerce.com (password hash seeded)
--   Staff:  owner@coolshoes.com, owner@techstore.com (password hashes seeded)
--   Customers Shop A: alice@example.com, bob@example.com (password hashes seeded)
--   Customers Shop B: alice@example.com (password hash seeded)
--     (same email different shops — proves tenant isolation)
--
-- Test Scenarios:
--   ✓ Staff login:    POST /api/v1/auth/staff/login
--   ✓ Customer login: POST /api/v1/auth/customer/login (needs X-Shop-ID)
--   ✓ Admin login:    POST /api/v1/auth/admin/login
--   ✓ Tenant isolation for customers across shops
-- ============================================================
