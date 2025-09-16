# API Documentation (Swagger / OpenAPI)

Nexus Commerce uses OpenAPI as the API contract source of truth for public portfolio documentation.

Backend API does not expose a Swagger endpoint. Internal API documentation is served from [openapi.yaml](openapi.yaml).

## Files

- OpenAPI spec: [openapi.yaml](openapi.yaml)

## Quick View with Swagger UI

Swagger UI should use the curated OpenAPI file only.

Run from repository root:

```powershell
docker run --rm -p 8081:8080 -e SWAGGER_JSON=/spec/openapi.yaml -v ${PWD}/docs:/spec swaggerapi/swagger-ui
```

Then open:
- http://localhost:8081

Or use the helper script:

```powershell
./scripts/swagger-ui.ps1
```

## Validate the OpenAPI Spec

Option 1 (Redocly CLI):

```powershell
npx @redocly/cli lint docs/openapi.yaml
```

Option 2 (Swagger CLI):

```powershell
npx swagger-cli validate docs/openapi.yaml
```

## Route Coverage Tracking

Route coverage is maintained directly in [openapi.yaml](openapi.yaml).

When adding or changing backend routes, update the matching path in [openapi.yaml](openapi.yaml) in the same change set.

## Scope and Conventions

- Tenant-scoped endpoints require `X-Shop-ID`.
- Auth uses JWT bearer tokens.
- Most API responses use a `data` envelope (with optional `meta`).
- Health/readiness endpoints may return plain status objects.

## Maintenance Rules

1. Update `docs/openapi.yaml` whenever backend routes or payloads change.
2. Keep examples realistic and aligned with current `seed.sql` defaults.
3. Prefer backward-compatible changes to schemas.
4. Remove deprecated fields/paths only with a migration note in release docs.
