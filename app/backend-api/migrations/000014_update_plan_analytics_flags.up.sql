-- Migrate plan features to granular analytics flags.
-- Replaces the single "advanced_analytics" flag with:
--   core_analytics        — basic visit counts, daily trend, live count
--   standard_analytics    — top pages, referrers, geo, device split
--   realtime_analytics    — SSE live stream, sparkline
--   advanced_analytics    — cohort, funnel, custom reports (premium only)
--   analytics_export      — CSV/JSON export
--   analytics_alerts      — threshold-based email alerts (premium only)
--   analytics_api_access  — programmatic API access (premium only)

UPDATE plans
SET features = '{"custom_domain":false,"support_automation":false,"coupons":true,"shipping_zones":true,"core_analytics":true}'
WHERE name = 'basic';

UPDATE plans
SET features = '{"custom_domain":true,"support_automation":false,"coupons":true,"shipping_zones":true,"core_analytics":true,"standard_analytics":true,"realtime_analytics":true,"analytics_export":true}'
WHERE name = 'professional';

UPDATE plans
SET features = '{"custom_domain":true,"support_automation":true,"coupons":true,"shipping_zones":true,"core_analytics":true,"standard_analytics":true,"realtime_analytics":true,"advanced_analytics":true,"analytics_export":true,"analytics_alerts":true,"analytics_api_access":true,"webhooks":true,"priority_support":true}'
WHERE name = 'premium';
