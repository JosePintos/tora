package server

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	handlers "github.com/JosePintos/tora/apps/api/internal/platform/server/handlers"

	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/generated"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/resolvers"
)

func NewRouter(resolver *resolvers.Resolver) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.HealthHandler)

	gqlServer := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	mux.Handle("/query", gqlServer)

	return mux
}
