package database

import (
	"context"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func New(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if err := godotenv.Load("../../../../../.env.test"); err != nil {
		t.Fatalf("failed to load .env.test: %v", err)
	}

	cfg := config.Load()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	Migrate(t, cfg.Database.URL)

	t.Cleanup(func() {
		pool.Close()
	})

	Cleanup(t, pool)

	return pool
}
