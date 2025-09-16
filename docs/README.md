# Nexus Commerce Documentation

Public documentation for the backend-first architecture of Nexus Commerce.

Frontend implementation details are intentionally minimized while the frontend is being rebuilt.

## Documentation Map

- API contract (Swagger/OpenAPI): [API.md](API.md)
- OpenAPI spec file: [openapi.yaml](openapi.yaml)
- Gateway architecture: [Gateway.md](Gateway.md)
- Infrastructure and operations: [infra.md](infra.md)
- Data model overview: [Schema.md](Schema.md)
- Feature flags and plan matrix: [Feature_Flags.md](Feature_Flags.md)
- AI engine overview: [AI_Engine.md](AI_Engine.md)

## Recommended Reading Order

1. [Gateway.md](Gateway.md)
2. [Schema.md](Schema.md)
3. [API.md](API.md)
4. [infra.md](infra.md)
5. [Feature_Flags.md](Feature_Flags.md)
6. [AI_Engine.md](AI_Engine.md)

## Source of Truth Rules

- Database schema source of truth: `app/backend-api/migrations`.
- Query contracts source of truth: `app/backend-api/sql/queries`.
- Regenerated query models: `app/backend-api/internal/db`.

Never manually edit files under `app/backend-api/internal/db`.