package server

import (
	"context"
	"net/http"

	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/resolvers"
	"github.com/JosePintos/tora/apps/api/internal/platform/config"
)

type Server struct {
	Config *config.Config
	Server *http.Server
}

func New(cfg *config.Config, resolver *resolvers.Resolver) *Server {
	router := NewRouter(resolver)
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}
	return &Server{
		Config: cfg,
		Server: server,
	}
}

func (s *Server) Run() error {
	return s.Server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Server.Shutdown(ctx)
}
