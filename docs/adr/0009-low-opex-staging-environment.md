# ADR 0009: On-Demand Webhook Staging Environment

Date: 2026-10-01
Status: Accepted

## Context

Validating external payment webhooks (Paymob) and destructive SQL migrations requires a publicly routable HTTPS endpoint, but running a duplicate production cluster continuously is unnecessary when integration tests run intermittently.

## Decision

Provision an on-demand hourly-billed VPS instance running the Docker Compose stack with reduced memory allocations (`shared_buffers = 256MB`) only during active webhook and migration verification windows.

## Consequences

- Provides an isolated public target for webhook signature and migration verification at minimal hourly compute cost.
- Requires provisioning automation to keep staging configuration aligned with `docker-compose.prod.yml`.
