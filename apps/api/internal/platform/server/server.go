package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/resolvers"
	"github.com/JosePintos/tora/apps/api/internal/platform/config"
)

type Server struct {
	Config *config.Config
	Server *http.Server
}

func New(cfg *config.Config, resolver *resolvers.Resolver, logger *slog.Logger) *Server {
	router := NewRouter(resolver, logger)
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second, // guards against Slowloris (slow-header) attacks
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
