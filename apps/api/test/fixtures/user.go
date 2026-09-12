package fixtures

import (
	"testing"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateUser(t *testing.T, queries *sqlc.Queries) sqlc.User {
	t.Helper()

	user, err := queries.CreateUser(
		t.Context(),
		sqlc.CreateUserParams{
			ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Username:     "testuser-" + uuid.NewString(),
			PasswordHash: "hashed_password",
			CreatedAt:    pgtype.Timestamptz{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		},
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return user
}
