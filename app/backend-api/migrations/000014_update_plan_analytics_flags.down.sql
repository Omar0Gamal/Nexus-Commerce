-- Revert plan features to pre-migration values.

UPDATE plans
SET features = '{"custom_domain":false,"support_automation":false,"coupons":true,"shipping_zones":true}'
WHERE name = 'basic';

UPDATE plans
SET features = '{"custom_domain":true,"support_automation":false,"coupons":true,"shipping_zones":true,"advanced_analytics":true}'
WHERE name = 'professional';

UPDATE plans
SET features = '{"custom_domain":true,"support_automation":true,"coupons":true,"shipping_zones":true,"advanced_analytics":true,"webhooks":true,"priority_support":true}'
WHERE name = 'premium';
