# Gateway Service

The gateway is the tenant-aware edge control plane for Nexus Commerce.

It sits between Caddy and the application stack and provides request normalization, tenant resolution, protection controls, and traffic telemetry.

## Responsibilities

- Resolve hostnames to tenant shops
- Inject trusted tenant context (`X-Shop-ID`)
- Apply per-IP and per-tenant rate limits
- Propagate distributed tracing (`X-Request-ID`)
- Collect traffic analytics and beacon events
- Proxy requests to backend and storefront upstreams

## Request Lifecycle

1. Request arrives from Caddy.
2. Global middleware runs (recovery, logging, tracing, CORS, instrumentation).
3. Tenant is resolved from host (subdomain or custom domain).
4. Request is rate-limited and tagged with tenant/tracing metadata.
5. Request is proxied:
   - `/api/*` to backend
   - non-API paths to storefront upstream

## Middleware Overview

Global middleware:
- panic recovery
- structured request logging
- tracing and request ID propagation
- instrumentation for metrics

Tenant-aware middleware:
- resolver (host -> tenant)
- rate limiter (shop-level fixed window)
- auth strict limiter for auth endpoints
- analytics collector middleware

## Exposed Endpoints

Operational:
- `GET /health`
- `GET /metrics`

Analytics:
- `POST /api/v1/analytics/beacon`
- `GET /beacon.js`

## Caching and Resolution

Tenant resolution uses three layers:
- L1: in-memory cache
- L2: Redis cache
- L3: PostgreSQL fallback

Invalidation events are propagated through Redis Pub/Sub.

## Key Environment Variables

- `PORT`
- `ROOT_DOMAIN`
- `BACKEND_URL`
- `FRONTEND_URL`
- `REDIS_ADDR`
- `DB_URL`
- `RATE_LIMIT_DEFAULT`

## Operational Guidance

- Prefer gateway URL (`:8000`) for local testing to exercise tenant/rate-limit middleware.
- Keep `GET /health` and `GET /metrics` integrated with monitoring.
- Verify request correlation by searching shared `X-Request-ID` in gateway and backend logs.

## Related Source

- `gateway/cmd/server/main.go`
- `gateway/internal/resolver/`
- `gateway/internal/ratelimit/`
- `gateway/internal/analytics/`
- `gateway/internal/proxy/`
