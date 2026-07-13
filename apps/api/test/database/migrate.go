package database

import (
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(t *testing.T, databaseURL string) {
	t.Helper()

	m, err := migrate.New(
		"file://../../../../../migrations",
		databaseURL,
	)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			t.Fatalf("failed closing migration source: %v", sourceErr)
		}
		if dbErr != nil {
			t.Fatalf("failed closing migration database: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to apply migrations: %v", err)
	}
}
