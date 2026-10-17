# ADR 0007: Standalone Next.js Container Behind Cloudflare Cache

Date: 2026-10-01
Status: Accepted

## Context

Customer storefronts require Server-Side Rendering (SSR) for search engine indexing and dynamic multi-tenant domain routing (`/[domain]`). Hosting multi-tenant SSR workloads on managed serverless platforms exposes the system to per-gigabyte bandwidth and invocation overages.

## Decision

Compile the Next.js application (`app/frontend`) with `output: "standalone"` and run it as a single Node 20 container (`frontend:3000`) inside Docker Compose behind Caddy and Cloudflare edge caching.

## Consequences

- Predictable fixed infrastructure cost alongside the Go API, PostgreSQL, and Redis containers.
- Single deployment serves `/admin`, `/superadmin`, and tenant storefronts via `src/middleware.ts` host inspection.
- Requires explicit `Cache-Control` headers and cache purge hooks when merchants mutate catalog data.
