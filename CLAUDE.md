# CLAUDE.md

This file provides guidance to Claude Code when working in this repository.

## Project

Tora is a Japanese-learning application built around spaced repetition. It is also a learning project for Go, GraphQL, Hexagonal Architecture, AWS, Terraform, and CI/CD (see `docs/vision.md`).

The MVP includes:
- User registration and login
- Deck management
- Vocabulary cards
- Study sessions
- Progress tracking

Project documentation:
- `docs/vision.md`: product vision and learning goals.
- `docs/domain.md`: domain concepts and business rules.
- `docs/architecture.md`: architectural guidelines.
- `docs/adrs/`: Architecture Decision Records. These document accepted architectural decisions.

Read relevant documentation before making architectural or domain decisions. Do not silently contradict an existing ADR. If a decision needs to change, explain why and propose an ADR update.

This is a monorepo managed with pnpm workspaces (`apps/*`, `packages/*`):

- `apps/api`: Go backend, a single deployable modular monolith exposing GraphQL.
- `apps/web`: React 19 + TypeScript + Vite frontend, currently the unmodified Vite template.

Most current implementation work is in `apps/api`; `apps/api/README.md` covers its setup.

Note: `apps/api/LICENSE` is golang-migrate's MIT license, copied in by mistake. It is not the project's license.

## Working principles

When making changes:

1. Inspect the relevant source files, tests, and documentation before editing.
2. Follow existing project conventions unless there is a concrete reason to change them.
3. Prefer the smallest coherent implementation that solves the problem.
4. Avoid unnecessary abstractions, dependencies, frameworks, and architectural complexity.
5. Do not introduce microservices. The intended architecture is a modular monolith.
6. Do not implement features outside the requested scope.
7. Do not silently change public APIs, domain behavior, database contracts, or architectural decisions.
8. Never hand-edit generated code.
9. Do not claim that tests or commands passed unless you actually ran them.
10. If requirements are ambiguous or an architectural decision has significant trade-offs, explain the options before proceeding.

Keep business logic independent of delivery and infrastructure technologies.

## Mentoring mode

This project exists so the owner grows as an engineer and can do well in job interviews. Act as a senior engineer pairing with them, not only as a code generator.

- **Implement what was asked, then teach.** Finish the requested task within the existing architecture. Raise improvements separately; don't fold them into the task.
- **Flag what you notice, with a severity.** Label each issue *bug*, *security*, *design* or *nit*, and point to the file and line. Fix bugs and security problems that are in scope. For anything else, explain it and let the owner decide.
- **Explain why.** For every recommendation, give the trade-off, at least one alternative, and when the current approach is fine. Name the concept (for example "interface segregation", "Slowloris", "N+1 queries") so the owner can study it further.
- **Connect to interviews when relevant.** If something is a common interview topic (system design, concurrency, security, testing strategy, SQL indexing), say so in one line.
- **Prefer current, idiomatic tooling** (recent Go and library versions, the standard library first) and say when a tool or API is deprecated.
- **Leave learning-sized work to the owner.** When a follow-up is a good exercise, describe it and offer to review it rather than writing it straight away.
- **Decisions can change.** If a recommendation contradicts an ADR or documented convention, propose an ADR update instead of silently diverging.

## API architecture

The API follows Domain-Driven Design (DDD) for feature organization and Hexagonal Architecture for dependency direction.

Each bounded context under `internal/` (`user`, `deck`, `card`) follows this structure:

```text
<context>/
├── domain/
├── application/
│   ├── commands/
│   ├── ports/
│   └── services/
└── infrastructure/
    └── persistence/
        └── postgres/
```

### Domain

`domain/` contains business entities, invariants, and domain behavior.

- Keep domain logic independent of GraphQL, PostgreSQL, sqlc, and HTTP.
- Use validating constructors such as `NewUser`, `NewDeck`, and `NewCard`.
- Constructors generate entity UUIDs and `CreatedAt` timestamps.
- Enforce business invariants in the domain, not in GraphQL resolvers or database adapters.
- Avoid importing other architectural layers into the domain.

### Application

`application/commands/` contains input DTOs for use cases, such as `CreateDeckCommand` and `LoginCommand`.

`application/ports/` contains interfaces defining the capabilities required by application services, such as repositories, password hashing, and token generation.

`application/services/` contains use cases. Services depend on domain objects and ports, not concrete infrastructure implementations.

**Go interface convention:** Store and accept interfaces by value, not by pointer. For example:

```go
type DeckService struct {
    deckRepository ports.DeckRepository
}

func NewDeckService(repo ports.DeckRepository) *DeckService {
    return &DeckService{
        deckRepository: repo,
    }
}
```

Avoid `*ports.DeckRepository`. A concrete pointer such as `*postgres.Repository` can implement and be assigned to the `ports.DeckRepository` interface.

### Infrastructure

`infrastructure/` contains outbound adapters implementing application ports.

`persistence/postgres/`:
- Implements repository interfaces.
- Uses generated sqlc queries.
- Contains persistence-to-domain mapping in `mapper.go`.
- Handles PostgreSQL-specific types such as `pgtype`.

`user/infrastructure/auth/` contains the bcrypt and JWT adapters.

Repository implementations should include compile-time interface assertions:

```go
var _ ports.DeckRepository = (*Repository)(nil)
```

### Dependency direction

Dependencies point inward:

```text
GraphQL / HTTP
      |
      v
Application services
      |
      v
Domain + application ports
      ^
      |
Infrastructure adapters

Bootstrap wires the implementations together.
```

- Inbound adapters call application services.
- Application services depend on ports.
- Outbound adapters implement those ports.
- Bootstrap is responsible for connecting concrete implementations.

Do not put business logic in GraphQL resolvers or infrastructure adapters.

### Repository conventions

When a requested row does not exist, repository methods currently return `(nil, nil)` rather than exposing `pgx.ErrNoRows`.

Application services are responsible for translating that result into appropriate errors, such as `ErrUserNotFound`.

Preserve this convention unless there is an explicit decision to change it. Be careful to distinguish a missing record from an actual database failure.

Prefer wrapped errors using `%w` and `errors.Is` when callers need to identify specific errors.

## Wiring and request flow

1. `cmd/server/main.go` loads `.env` using godotenv and calls `platform/app.New()`.
2. `platform/app` loads configuration, opens the pgx connection pool, and builds the application container.
3. `platform/bootstrap` is the single composition root. It creates concrete adapters and application services and exposes them through a `Container`. The package is declared as `bootstrap` and imported under that alias.
4. `interfaces/graphql/resolvers.Resolver` holds the container.
5. Resolvers convert GraphQL inputs into application commands, invoke services, and map results into GraphQL models.
6. Domain-to-GraphQL conversion lives in `interfaces/graphql/mapper`.
7. `platform/server/router.go` mounts `/query`, wrapped in `interfaces/graphql/auth.AuthMiddleware`.

Keep resolvers thin. They should handle delivery concerns, not implement business rules.

## GraphQL

GraphQL is centralized under `internal/interfaces/graphql/`.

- `schema/` contains `.graphqls` schema files.
- `resolvers/` contains resolver implementations.
- `model/` contains generated GraphQL models.
- `generated/` contains gqlgen execution code.
- `mapper/` contains conversions between domain objects and GraphQL models.
- `auth/` contains authentication middleware and context helpers.

Do not expose domain entities or sqlc-generated database models directly through GraphQL.

GraphQL schema files share one executable schema.

- `type Query` is declared once in `health.graphqls`.
- `type Mutation` is declared once in `user.graphqls`.
- Other schema files extend these root types using `extend type`.

## Authentication

- Login returns an HS256 JWT.
- The JWT `sub` claim contains the user UUID.
- Tokens expire after 24 hours.
- `JWT_SECRET` is loaded from environment configuration.
- Never hardcode production secrets or commit real secrets.
- Passwords must be hashed using bcrypt. Never store or return plaintext passwords.

`AuthMiddleware` does not reject requests. A missing, malformed, or invalid `Authorization: Bearer` header means the request continues without an authenticated user ID in its context.

Protected resolvers must explicitly call:

```go
userID, ok := auth.UserIDFromContext(ctx)
if !ok {
    return nil, errors.New("unauthenticated")
}
```

Do not assume the middleware enforces authorization.

Ownership must come from the authenticated context, never from client-supplied ownership fields. For example, deck creation obtains `OwnerID` from the authenticated user rather than accepting it in `CreateDeckInput`.

## Database and migrations

Migrations are version-controlled schema changes.

- Migration files follow `00000N_name.{up,down}.sql`.
- Every migration must have both an up and down file.
- Down migrations use `DROP TABLE IF EXISTS` where appropriate.
- Do not modify migrations that have already been applied to shared or persistent databases. Create a new migration instead.
- Include appropriate foreign keys, constraints, and indexes.
- Preserve referential integrity between bounded contexts at the persistence boundary.

The database schema is defined by migrations. sqlc uses migrations as its schema source and SQL queries from `sql/queries/`.

The server does not apply migrations at startup. Apply them explicitly before running the application.

## Commands

Run API commands from `apps/api`.

Development tools (`sqlc`, `gqlgen`, `migrate`, `air`) are installed globally in `~/go/bin`; they are not tracked in `go.mod` as Go tools.

```bash
go build ./...
go vet ./...
air
go run ./cmd/server
go test ./...
```

Other useful commands:

```bash
# Run a specific test
go test ./internal/user/application/services -run TestLogin_WrongPassword

# Generate sqlc code after editing migrations or SQL queries
sqlc generate

# Generate gqlgen code after editing GraphQL schemas
gqlgen generate
```

### Databases

Databases are defined in the root `docker-compose.yml`:

- `postgres`: development database, port `5432`, database `tora`.
- `postgres-test`: integration-test database, port `5433`, database `tora_test`.

Start both from the repository root:

```bash
docker compose up -d postgres postgres-test
```

### Migrations

The development database must be migrated before starting the server.

```bash
migrate -path migrations -database "$DATABASE_URL" up
```

The `migrate` CLI does not automatically load `.env`. Ensure `DATABASE_URL` is exported in the shell before running this command, or provide the connection string directly.

Migrations can also be run using the `migrate` compose service under the `tools` profile.

### Environment

`apps/api/.env` contains development configuration:

- `APP_ENV`
- `PORT` (defaults to `8080`)
- `DATABASE_URL`
- `JWT_SECRET`

The server refuses to start if `JWT_SECRET` is empty.

Integration tests use `apps/api/.env.test`.

Do not commit actual secrets or credentials.

### Frontend

Run web commands from the repository root:

```bash
pnpm dev:web
pnpm build:web
pnpm lint:web
```

## Code generation

Never hand-edit generated files. Modify their source files and regenerate.

| Source | Command | Generated output |
|---|---|---|
| `migrations/*.sql` and `sql/queries/*.sql` | `sqlc generate` | `internal/platform/database/sqlc/` |
| `internal/interfaces/graphql/schema/*.graphqls` | `gqlgen generate` | `graphql/generated/generated.go` and `graphql/model/models_gen.go` |

gqlgen uses the follow-schema resolver layout: `foo.graphqls` maps to `resolvers/foo.resolvers.go`.

gqlgen generally preserves resolver implementations when regenerating, but always inspect the resulting diff. Never assume generated changes are harmless.

## Testing

Testing is organized by architectural layer.

### Unit tests

- Domain tests verify entity invariants and behavior.
- Application service tests use hand-written repository mocks.
- Resolver tests verify GraphQL-to-application behavior.
- Middleware tests verify request authentication and context behavior.

Mocks use optional `XxxFn` function fields (for example `CreateFn` or `GenerateTokenFn`) and are defined in their respective test files. There is no mocking library. Some older tests use `XxxFunc`; use `Fn` in new code.

Do not use a real PostgreSQL database for domain, service, resolver, or middleware unit tests.

### Repository integration tests

Repository integration tests live under:

```text
*/infrastructure/persistence/postgres/repository_test.go
```

They require `postgres-test` to be running.

Tests use `test/database.New(t)`, which:
- Loads `.env.test`.
- Runs migrations.
- Cleans up test data by truncating `users ... CASCADE`.

Shared fixtures live in `test/fixtures`.

**Current limitation:** the database test helper locates `.env.test` and `migrations/` through relative paths and currently assumes test packages are exactly five directory levels below `apps/api`.

Do not add integration-test packages at a different nesting level without adapting and testing the helper.

Run all tests, including integration tests:

```bash
go test ./...
```

Run unit tests without PostgreSQL (Bash/Git Bash):

```bash
go test $(go list ./... | grep -v persistence/postgres)
```

## Feature status

The `card` context currently has:
- Domain implementation.
- Repository port.
- Migration `000004`.
- sqlc queries.

The application service, PostgreSQL repository adapter, and GraphQL operations are not yet implemented. Verify the actual code before assuming this status is still current, and update this section as the feature progresses.

### Known gaps

These are acknowledged, not yet addressed. Don't fix them as a side effect of an unrelated task, but mention one when it is relevant to the work at hand. Remove an entry once it is resolved.

- Errors reach GraphQL clients as raw strings: there is no error presenter or `extensions.code`, and database errors such as a duplicate username leak through.
- Login reveals whether a username exists, both through different error messages and through timing (bcrypt only runs for existing users).
- The server uses the deprecated `handler.NewDefaultServer`, with no query complexity or depth limit and introspection always on.
- `decks.owner_id` has no index, although `ListDecksByOwner` filters on it.
- Validation is split: username and password rules live in `UserService`, and `DeckService` repeats the domain's empty-name check.
- `DeckRepository` declares `Update` and `Delete`, whose Postgres adapter panics with "not implemented".
- The card port takes `deckID *uuid.UUID`, while everything else uses `uuid.UUID`.
- The domain calls `time.Now()` directly, so timestamps can't be controlled in tests.

## Repository conventions

- Work on a feature branch.
- This is a solo project; pull requests are optional.
- Merge into `main` using `git merge --no-ff` when the work is ready.
- Do not push, merge, or rewrite Git history unless explicitly requested.
- There is currently no CI pipeline; `.github/` is empty.

## Definition of done

Before considering a change complete:

1. Ensure the implementation follows the existing architecture and conventions.
2. Run `gofmt` on modified Go files.
3. Run relevant unit tests.
4. Run integration tests when database-related behavior changes and PostgreSQL is available.
5. Run `go vet ./...` when practical.
6. Inspect generated files and the final Git diff.
7. Report:
   - What changed.
   - Important design decisions.
   - Tests and commands actually executed.
   - Any remaining limitations or follow-up work.