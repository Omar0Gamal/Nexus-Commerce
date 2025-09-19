# Infrastructure and Operations

This document covers how Nexus Commerce runs in development and production.

## Environment Profiles

## Development (`docker-compose.dev.yml`)

Primary goal: fast iteration.

Characteristics:
- hot reload for backend and gateway
- local debugging ports for Postgres and Redis
- one-off migration job via compose profile

Core services:
- `postgres`
- `redis`
- `backend`
- `gateway`
- `frontend` (currently secondary while frontend is under rewrite)
- `migrate`

Recommended bootstrap:

```powershell
./scripts/rebuild.ps1 -Full
```

## Production (`docker-compose.prod.yml`)

Primary goal: stability and operational visibility.

Core services:
- edge: `caddy`
- app: `gateway`, `backend`, `frontend`
- data: `postgres`, `redis`
- ops: `backup`, `loki`, `promtail`, `grafana`, `portainer`

## Traffic Topology

1. Public traffic enters Caddy.
2. Caddy forwards to gateway.
3. Gateway resolves tenant and routes request class.
4. Backend serves API operations against Postgres/Redis.

## Required Configuration

Minimum production environment variables:
- `DB_PASSWORD`
- `ROOT_DOMAIN`
- `GRAFANA_PASSWORD`

Also configure integration credentials as needed:
- object storage
- payment provider
- SMTP mailer
- AI provider/engine keys

Start from `.env.example` and extend per environment.

## Persistence

Persistent volumes are used for:
- Postgres data
- Redis data
- Grafana state
- Loki state
- backups
- Caddy runtime data

## Health, Logs, and Metrics

- Health checks are configured in compose for core services.
- Metrics are exposed by gateway/backend and visualized in Grafana.
- Logs flow through Promtail to Loki.

## Operational Commands

Start production stack:

```powershell
docker compose -f docker-compose.prod.yml up -d --build
```

Run migrations:

```powershell
docker compose -f docker-compose.prod.yml run --rm migrate
```

Check service state:

```powershell
docker compose -f docker-compose.prod.yml ps
```

## Production Readiness Checklist

- secrets are stored outside git
- migration job succeeds on target environment
- health and metrics endpoints are monitored
- backup path is verified
- gateway and backend logs include request correlation IDs
