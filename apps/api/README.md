# Tora API

Go backend for Tora: a modular monolith that exposes a single GraphQL endpoint. It is organised as DDD bounded contexts (`internal/user`, `internal/deck`, `internal/card`) using hexagonal architecture. Background and decisions live in [`docs/`](../../docs), and contributor and agent conventions in [`CLAUDE.md`](../../CLAUDE.md).

## Stack

| Concern | Tool |
|---|---|
| HTTP + GraphQL | `net/http`, [gqlgen](https://gqlgen.com) (schema-first) |
| Database | PostgreSQL 17, [pgx v5](https://github.com/jackc/pgx) |
| Queries | [sqlc](https://sqlc.dev) (SQL in, typed Go out) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Auth | bcrypt password hashing, HS256 JWT |
| Hot reload | [air](https://github.com/air-verse/air) |

## Getting started

Prerequisites: Go (see `go.mod`), Docker, and the CLIs `sqlc`, `gqlgen`, `migrate` and `air` installed with `go install`.

1. Start the databases (from the repo root):

   ```bash
   docker compose up -d postgres postgres-test
   ```

2. Create `apps/api/.env`:

   ```dotenv
   APP_ENV=development
   PORT=8080
   DATABASE_URL=postgres://postgres:postgres@localhost:5432/tora?sslmode=disable
   JWT_SECRET=<any long random string>
   ```

3. Apply the migrations:

   ```bash
   migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/tora?sslmode=disable" up
   ```

4. Run the server with hot reload (or `go run ./cmd/server`):

   ```bash
   air
   ```

GraphQL is served at `POST http://localhost:8080/query`. To call authenticated operations, send `Authorization: Bearer <token>` with the token returned by the `login` mutation.

## Tests

```bash
go test ./...                                                # unit + integration (needs postgres-test)
go test $(go list ./... | grep -v persistence/postgres)      # unit only
```

Integration tests read `.env.test` and apply the migrations to `tora_test` themselves.

## Code generation

```bash
sqlc generate     # after changing migrations/ or sql/queries/
gqlgen generate   # after changing internal/interfaces/graphql/schema/
```
