# Data Model Overview

This is a high-level schema map for Nexus Commerce.

Authoritative sources:
- migrations: `app/backend-api/migrations`
- query contracts: `app/backend-api/sql/queries`

Use this file for architecture understanding, not as executable SQL.

## Multi-Tenancy Model

Nexus Commerce is tenant-isolated by `shop_id`.

Rules:
- tenant-owned records are scoped by `shop_id`
- gateway resolves host to tenant and injects `X-Shop-ID`
- backend handlers/services/queries enforce shop scope

## Domain Areas

## Platform and Billing

Representative tables:
- `plans`
- `shops`
- `shop_invoices`
- `platform_admins`

Concerns:
- subscription plans and limits
- shop lifecycle and billing state
- platform-level administration

## Identity and Access

Representative tables:
- `users`
- `customers`
- `shop_staff`
- `shop_roles`
- `user_secrets`

Concerns:
- multi-actor authentication
- role and permission model
- 2FA and account security state

## Catalog and Commerce

Representative tables:
- `categories`
- `products`
- `product_variants`
- `product_images`
- `orders`
- `order_items`
- `returns`
- `coupons`
- `promotions`

Concerns:
- product lifecycle
- inventory and pricing
- checkout and order fulfillment

## Customer Engagement

Representative tables:
- `wishlists`
- `product_reviews`
- `social_proof_events`
- notification preference/queue tables

Concerns:
- engagement loops
- trust/reputation systems
- outbound communication signals

## Integrations and Automation

Representative tables:
- webhook configuration and delivery metadata
- payment configuration/credentials
- background job payload/execution data

Concerns:
- external integrations
- asynchronous workflows
- operational event delivery

## Analytics and AI

Representative tables are introduced by newer migrations and evolve over time.

Concerns:
- analytics aggregation and retention
- AI configuration and usage controls

## Migration and sqlc Workflow

1. Add schema change migration (`.up.sql` and `.down.sql`).
2. Add or update SQL query contracts under `sql/queries`.
3. Run `sqlc generate`.
4. Consume generated query code in service/handler layers.

Never manually edit generated files under `app/backend-api/internal/db`.

## Useful Commands

Run migrations in development:

```powershell
docker compose -f docker-compose.dev.yml run --rm migrate
```

Regenerate sqlc artifacts:

```powershell
Set-Location app/backend-api
sqlc generate
```
