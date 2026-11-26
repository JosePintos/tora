package server

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	handlers "github.com/JosePintos/tora/apps/api/internal/platform/server/handlers"

	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/auth"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/errorpresenter"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/generated"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/loaders"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/resolvers"
)

func NewRouter(resolver *resolvers.Resolver, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.HealthHandler)

	gqlServer := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	gqlServer.SetErrorPresenter(errorpresenter.New(logger))
	gqlServer.SetRecoverFunc(errorpresenter.Recover(logger))

	// Order matters: AuthMiddleware puts the user ID in the context, and the
	// loaders middleware needs it to build per-user, per-request DataLoaders.
	mux.Handle("/query", auth.AuthMiddleware(resolver.Container.TokenProvider,
		loaders.Middleware(resolver.Container.CardService, gqlServer)))

	return mux
}
