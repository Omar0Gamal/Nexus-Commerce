# Feature Flags and Plan Access

This document summarizes backend plan feature flags used for capability gating.

Source of truth:
- `app/backend-api/seed.sql` (`plans.features` JSON)
- runtime middleware checks via billing feature gates

Legend:
- `Y` = enabled
- `N` = disabled

## Plans

- `basic`
- `professional`
- `premium`

## Core Platform Features

| Feature Key | Basic | Professional | Premium |
| --- | --- | --- | --- |
| `custom_domain` | N | Y | Y |
| `coupons` | Y | Y | Y |
| `shipping_zones` | Y | Y | Y |
| `webhooks` | N | Y | Y |
| `support_automation` | N | N | Y |
| `priority_support` | N | N | Y |

## Analytics Features

| Feature Key | Basic | Professional | Premium |
| --- | --- | --- | --- |
| `core_analytics` | Y | Y | Y |
| `realtime_analytics` | N | Y | Y |
| `standard_analytics` | N | Y | Y |
| `advanced_analytics` | N | N | Y |
| `analytics_export` | N | Y | Y |
| `analytics_alerts` | N | N | Y |
| `analytics_api_access` | N | N | Y |

## SEO Features

| Feature Key | Basic | Professional | Premium |
| --- | --- | --- | --- |
| `seo_basic` | Y | Y | Y |
| `seo_scoring` | N | Y | Y |
| `seo_redirects` | N | Y | Y |
| `seo_advanced` | N | N | Y |
| `seo_search_console` | N | N | Y |

## AI Features

| Feature Key | Basic | Professional | Premium |
| --- | --- | --- | --- |
| `ai_text` | N | Y | Y |
| `ai_seo` | N | Y | Y |
| `ai_translate` | N | Y | Y |
| `ai_email` | N | Y | Y |
| `ai_reports` | N | N | Y |
| `ai_engine` | N | Y | Y |
| `ai_vision` | N | N | Y |
| `ai_forecasting` | N | N | Y |
| `ai_churn_scoring` | N | N | Y |

## Commerce and Engagement Extensions

| Feature Key | Basic | Professional | Premium |
| --- | --- | --- | --- |
| `wishlists` | Y | Y | Y |
| `product_reviews` | Y | Y | Y |
| `review_moderation` | N | Y | Y |
| `social_proof` | N | Y | Y |
| `abandoned_cart_recovery` | N | Y | Y |
| `multi_currency` | N | Y | Y |
| `multi_warehouse` | N | N | Y |
| `loyalty_program` | N | N | Y |
| `referral_program` | N | N | Y |

## Change Control Rules

1. Add or remove flags in migrations/seed flow first.
2. Keep this file synchronized with `seed.sql`.
3. Gate routes explicitly in backend middleware/handlers.
4. Keep frontend-only gating concerns out of this file while frontend is in rewrite.
