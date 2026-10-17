# ADR 0004: Compile-Time SQL Generation via sqlc

Date: 2026-09-30
Status: Accepted

## Context

Database access in Go requires balancing query control with type safety. Runtime ORMs obscure SQL execution plans and defer column mismatches to runtime, while manual `database/sql` scanning introduces repetitive boilerplate.

## Decision

Use `sqlc` with `pgx/v5` (`app/backend-api/sqlc.yaml`). Queries are authored in raw SQL (`app/backend-api/sql/queries/*.sql`) against the schema definitions (`app/backend-api/sql/schema/*.sql`) and compiled into strongly typed Go methods in `app/backend-api/internal/db/`.

## Consequences

- Explicit control over PostgreSQL queries, indexes, and `tsvector` full-text search operations.
- Schema and query parameter mismatches fail at compile time.
- Schema or query edits require running `sqlc generate`.
