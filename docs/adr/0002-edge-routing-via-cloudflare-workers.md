# ADR 0002: Edge Routing via Cloudflare Workers

Date: 2026-09-30
Status: Accepted

## Context

Multi-tenant requests arrive on platform subdomains (`{shop}.egyptcommerce.com`) or merchant custom domains (`www.kicksandco.store`) and must resolve to a tenant UUID (`shop_id`) before reaching the application origin. Operating a custom Go reverse proxy for custom-domain TLS termination and tenant lookup adds certificate management complexity and origin CPU overhead.

## Decision

Terminate TLS and resolve tenant hostnames at the network edge using Cloudflare Workers (`cloudflare/tenant-router/src/index.js`) backed by Cloudflare KV (`TENANT_KV`) and Cloudflare for SaaS. The Worker injects `X-Shop-ID` and `X-Gateway-Secret` headers before forwarding requests to the Caddy origin proxy.

## Consequences

- Custom-domain TLS issuance and renewal are offloaded to Cloudflare for SaaS.
- Origin servers verify `X-Gateway-Secret` in Caddy and `TenantMiddleware` to reject direct header spoofing.
- Edge routing depends on Cloudflare Workers and KV availability.
