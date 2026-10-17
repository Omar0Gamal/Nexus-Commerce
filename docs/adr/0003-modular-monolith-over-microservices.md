# ADR 0003: Modular Monolith over Microservices

Date: 2026-09-30
Status: Accepted

## Context

The backend spans catalog, inventory, cart, checkout, orders, billing, analytics, and notifications. Splitting these domains into separate network services introduces RPC serialization overhead, distributed transaction failure modes, and multi-container orchestration costs.

## Decision

Compile all backend domains into a single Go 1.24 binary (`app/backend-api/cmd/api`) structured into isolated internal packages (`app/backend-api/internal/*`). Cross-domain operations execute via in-process function calls and shared PostgreSQL transactions wired in `cmd/api/services.go`.

## Consequences

- Single static binary deployment (`gcr.io/distroless/static-debian12`) with zero inter-service network latency.
- Cross-domain writes (such as order capture and stock deduction) execute atomically inside one database transaction.
- Package boundaries under `internal/` must be enforced during code review to prevent cyclic dependencies.
