# Nexus Commerce

Multi-tenant e-commerce SaaS platform built with Go, Next.js, and a dedicated AI Engine.

Nexus Commerce is designed as a production-style portfolio project that combines:
- A tenant-aware edge gateway
- A modular backend API
- A storefront/dashboard frontend (currently being rebuilt)
- An optional AI microservice for ML workloads

## Why This Project

Most e-commerce demos only show a storefront.
Nexus Commerce demonstrates full platform engineering:
- Merchant operations dashboard
- Platform admin controls
- Multi-tenancy and domain routing
- Rate limiting, caching, observability, and background jobs

## Architecture

```text
Internet
	|
	v
Caddy (edge reverse proxy)
	|
	v
Gateway (Go)
  - tenant resolution (host -> shop)
  - rate limiting
  - request tracing
  - analytics beacon
  - smart routing
  |-----------------------------> Frontend (currently in rewrite)
	|
	v
Backend API (Go + Gin + sqlc)
  - auth / RBAC / billing
  - catalog / orders / cart / shipping
  - coupons / promotions / reviews
  - webhooks / notifications / analytics
  - worker queue
	|
	+--> PostgreSQL 16
	+--> Redis 7
	+--> Object Storage (S3 / Cloudflare R2)
	+--> AI Engine (FastAPI, optional)
```

## Core Capabilities

### Merchant and Platform
- Staff auth, customer auth, and platform-admin auth flows
- Product, category, variant, inventory, and media management
- Orders, returns, shipping zones/rates, and customer management
- Coupons, promotions, billing plans, and feature gating
- Platform admin pages for shop and plan oversight

### Storefront
- Frontend implementation is being rebuilt
- Backend APIs for storefront flows remain available and documented via OpenAPI

### Engineering Features
- Multi-tenant domain/subdomain routing
- Tenant-scoped API isolation via `X-Shop-ID`
- Redis-backed background jobs and asynchronous workflows
- Prometheus metrics, request tracing, and structured logs
- CI workflow for backend and gateway (`go vet`, `go build`, `go test`)

## Tech Stack

| Layer | Stack |
| --- | --- |
| Frontend | Next.js 15, React 19, TypeScript, Tailwind CSS (rewrite in progress) |
| Backend API | Go 1.24, Gin, pgx, sqlc, Redis |
| Gateway | Go 1.24, net/http, Redis, pgx |
| AI Engine | FastAPI, Python, model registry + inference routes |
| Data | PostgreSQL 16, Redis 7 |
| Infra | Docker Compose, Caddy, Grafana, Loki, Promtail, Portainer |

## Repository Layout

```text
app/
  backend-api/   # Go modular backend API
  frontend/      # Next.js storefront + dashboards
  ai-engine/     # Python FastAPI AI service
gateway/         # Tenant-aware reverse proxy in Go
docs/            # Architecture and implementation docs
config/          # Observability and infra config
scripts/         # Utility scripts (rebuild/reseed)
```

## Local Development

### Prerequisites
- Docker + Docker Compose
- PowerShell (for helper scripts on Windows)

### 1) Configure environment

```powershell
Copy-Item .env.example .env
```

At minimum, set a strong `DB_PASSWORD` in `.env`.

### 2) Quick start (recommended)

```powershell
./scripts/rebuild.ps1 -Full
```

This script:
- Rebuilds services
- Starts core containers
- Runs migrations
- Seeds test data

### 3) Manual start (alternative)

```powershell
docker compose -f docker-compose.dev.yml up -d postgres redis backend gateway frontend
docker compose -f docker-compose.dev.yml run --rm migrate
```

### Local URLs
- Gateway: http://localhost:8000
- Backend API: http://localhost:8080
- Frontend: http://localhost:3000

## Testing and Quality

Run tests locally:

```powershell
Set-Location app/backend-api; go test ./...
Set-Location ..\..\gateway; go test ./...
```

Current CI workflow is in `.github/workflows/ci.yml` and runs:
- `go vet ./...`
- `go build ./...`
- `go test ./...`

## Documentation

- Documentation index: [docs/README.md](docs/README.md)
- API docs (Swagger/OpenAPI): [docs/API.md](docs/API.md)
- OpenAPI spec: [docs/openapi.yaml](docs/openapi.yaml)
- Swagger UI helper script: [scripts/swagger-ui.ps1](scripts/swagger-ui.ps1)
- Gateway architecture: [docs/Gateway.md](docs/Gateway.md)
- AI Engine specification: [docs/AI_Engine.md](docs/AI_Engine.md)
- Feature plan matrix: [docs/Feature_Flags.md](docs/Feature_Flags.md)
- Data model: [docs/Schema.md](docs/Schema.md)

## Production Notes

Production stack is defined in `docker-compose.prod.yml` and includes:
- Caddy reverse proxy
- Gateway + Backend + Frontend
- PostgreSQL + Redis
- Backup, Grafana, Loki, Promtail, Portainer

Example:

```powershell
docker compose -f docker-compose.prod.yml up -d --build
docker compose -f docker-compose.prod.yml run --rm migrate
```

## Portfolio Highlights

This repo showcases:
- Distributed system design in a single repository
- Practical multi-tenancy and request routing
- Type-safe SQL workflow with sqlc
- Security-minded patterns (RBAC, token auth, rate limiting)
- Real operational concerns (health checks, metrics, logs, backups)

## Current Gaps (Honest Assessment)

- Some backend modules still need deeper unit/integration test coverage
- Frontend rewrite is in progress and intentionally not the current focus

## License

Proprietary - All rights reserved.

This repository is published as a study case and portfolio project.
Commercial use is reserved exclusively to the repository owner.
No permission is granted to third parties to use, copy, modify, deploy,
or redistribute this code without prior written permission.

See [LICENSE](LICENSE) for details.
