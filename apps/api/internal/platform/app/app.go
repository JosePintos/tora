package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/resolvers"
	bootstrap "github.com/JosePintos/tora/apps/api/internal/platform/bootstrap"
	"github.com/JosePintos/tora/apps/api/internal/platform/config"
	"github.com/JosePintos/tora/apps/api/internal/platform/database"
	"github.com/JosePintos/tora/apps/api/internal/platform/logger"
	"github.com/JosePintos/tora/apps/api/internal/platform/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config    *config.Config
	Server    *server.Server
	Logger    *slog.Logger
	Database  *pgxpool.Pool
	Container *bootstrap.Container
}

func New() *App {
	cfg := config.Load()
	log := logger.New()

	app := &App{
		Config: cfg,
		Logger: log,
	}

	db, err := app.initializeDatabase()
	if err != nil {
		app.Logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	app.Database = db
	app.Container = bootstrap.New(db)

	resolver := resolvers.New(app.Container)
	app.Server = server.New(cfg, resolver)

	return app
}

func (a *App) Run() {
	a.Logger.Info("starting server", "port", a.Config.Server.Port, "environment", a.Config.Environment)

	go func() {
		if err := a.Server.Run(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("server failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.Logger.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Logger.Error("server shutdown failed", "error", err)
	}

	a.Database.Close()
	a.Logger.Info("database connection closed")

	a.Logger.Info("application stopped")
}

func (a *App) initializeDatabase() (*pgxpool.Pool, error) {
	db, err := database.New(a.Config)
	if err != nil {
		a.Logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	return db, nil
}
