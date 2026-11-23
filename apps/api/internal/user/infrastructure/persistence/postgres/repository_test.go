package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
	testdb "github.com/JosePintos/tora/apps/api/test/database"
)

func TestCreateUser(t *testing.T) {
	db := testdb.New(t)

	queries := sqlc.New(db)

	repo := New(queries)

	user, err := domain.NewUser("testuser", "hashed_password")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	savedUser, err := queries.GetUserByUsername(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("failed to get user by username: %v", err)
	}

	if savedUser.Username != user.Username {
		t.Errorf("expected username %s, got %s", user.Username, savedUser.Username)
	}

	if savedUser.PasswordHash != user.PasswordHash {
		t.Errorf("expected password hash %s, got %s", user.PasswordHash, savedUser.PasswordHash)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	db := testdb.New(t)

	repo := New(sqlc.New(db))

	first, err := domain.NewUser("duplicate-user", "hashed_password")
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}
	if err := repo.Create(context.Background(), first); err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	second, err := domain.NewUser("duplicate-user", "other_hash")
	if err != nil {
		t.Fatalf("failed to build user: %v", err)
	}

	err = repo.Create(context.Background(), second)
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}
