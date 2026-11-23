# error-kinds

### Decision

Classify errors by kind with a tiny shared package, `internal/apperror`. The domain and application layers attach a kind (`Invalid`, `NotFound`, `Conflict`, `Unauthenticated`, `Forbidden`, `Internal`) to every error a client can act on. A single GraphQL error presenter maps kinds to `extensions.code`. Any error without a kind is treated as `Internal`: it is logged and its message is replaced with "internal server error".

### Status

Accepted

### Reason

Before this, resolvers returned `err.Error()` to clients, which caused three problems:
- Database details leaked, for example the name of the unique constraint on a duplicate username.
- Clients had to compare message strings to react to errors.
- Client mistakes and server failures looked the same, so there was no basis for deciding what to log or alert on.

With kinds:
- The domain decides what kind of failure something is, which is a business decision.
- Adapters translate vendor errors into domain errors.
- The edge translates kinds into its own transport vocabulary.
- Unknown errors are hidden by default, so forgetting to classify an error fails safe.

The presenter only knows six kinds, not every error in every bounded context, so adding a context or an error doesn't require changing it.

### Alternatives

- **One lookup table in the presenter** (`errors.Is(err, X)` mapped to a code). Simpler, but the GraphQL layer would have to import the errors of every bounded context and be edited for each new error. An error missing from the table silently becomes "internal".
- **Typed errors per context with an interface** (for example `interface{ Code() string }`). This avoids a shared package, but lets each context invent its own codes, and the codes are a transport concern that doesn't belong in the domain.
- **Returning `*gqlerror.Error` from resolvers.** This couples the delivery layer's error format to every resolver, and errors from services would still arrive unclassified.

### Consequences

- Every bounded context imports `internal/apperror`. Keep it free of dependencies on the rest of the module.
- Messages of classified errors are shown to clients, so they must not contain internal details.
- The `extensions.code` values are a public contract.
