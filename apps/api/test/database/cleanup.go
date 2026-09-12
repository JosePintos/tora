package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Cleanup(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	t.Cleanup(func() {
		_, err := db.Exec(
			context.Background(),
			`
			TRUNCATE TABLE users
			RESTART IDENTITY
			CASCADE;
			`,
		)

		if err != nil {
			t.Fatalf("failed cleaning database: %v", err)
		}
	})
}
