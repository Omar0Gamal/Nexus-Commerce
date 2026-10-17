# ADR 0008: First-Party Analytics Beacon Ingestion

Date: 2026-10-01
Status: Accepted

## Context

Storefronts must emit page-view, product-view, cart, and purchase telemetry to the backend analytics pipeline without relying on third-party SaaS trackers or coupling telemetry strictly to React client bundles.

## Decision

Ingest storefront events directly via `POST /api/v1/analytics/ingest` from a first-party tracker (`AnalyticsTracker.tsx` / `nexus-tracker.js`). The tracker maintains a client `session_id`, captures attribution parameters (`_fbp`, `_fbc`, `gclid`), and posts event batches to the Go API, which appends them to Redis Streams for asynchronous aggregation into PostgreSQL.

## Consequences

- Zero third-party telemetry licensing or egress fees.
- Full control over tenant data isolation and GDPR consent enforcement (`consent_log`).
- First-party ingestion endpoints must enforce rate limiting and payload validation.
