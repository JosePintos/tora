package services

import (
	"context"
	"errors"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/user/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
	"github.com/google/uuid"
)

type MockUserRepository struct {
	CreateFn    func(context.Context, *domain.User) error
	CreatedUser *domain.User
	CreateCalls int
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.CreateCalls++
	m.CreatedUser = user
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	panic("not implemented")
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	panic("not implemented")
}

type MockPasswordHasher struct {
	HashFn func(string) (string, error)
}

func (m MockPasswordHasher) Hash(password string) (string, error) {
	return m.HashFn(password)
}

func (MockPasswordHasher) Compare(hash, password string) error {
	return nil
}

func newTestUserService(repo ports.UserRepository) *UserService {
	hasher := MockPasswordHasher{
		HashFn: func(password string) (string, error) {
			return "hashed_password", nil
		},
	}

	return NewUserService(repo, hasher)
}

func TestCreateUser(t *testing.T) {
	repo := &MockUserRepository{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := newTestUserService(repo)

	user, err := service.Create(context.Background(), commands.CreateUserCommand{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.CreatedUser == nil {
		t.Fatal("expected user to be created in repository")
	}
	if repo.CreatedUser.Username != "testuser" {
		t.Fatalf("expected username to be testuser, got %s", repo.CreatedUser.Username)
	}
	if repo.CreatedUser.ID == uuid.Nil {
		t.Fatal("expected ID to be generated")
	}
	if repo.CreatedUser.PasswordHash != "hashed_password" {
		t.Fatal("expected password to be hashed")
	}
	if user == nil {
		t.Fatal("expected user to be returned")
	}
}

func TestCreateUser_ReturnsRepositoryError(t *testing.T) {
	repo := &MockUserRepository{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return errors.New("database error")
		},
	}

	service := newTestUserService(repo)
	user, err := service.Create(context.Background(), commands.CreateUserCommand{
		Username: "testuser",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if user != nil {
		t.Fatal("expected user to be nil on error")
	}
}

func TestCreateUser_ReturnsHashingError(t *testing.T) {
	hasher := MockPasswordHasher{
		HashFn: func(password string) (string, error) {
			return "", errors.New("hashing error")
		},
	}

	repo := &MockUserRepository{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := NewUserService(repo, hasher)
	_, err := service.Create(context.Background(), commands.CreateUserCommand{
		Username: "testuser",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repo.CreatedUser != nil {
		t.Fatal("repository should not be called when hashing fails")
	}
}

func TestCreateUser_EmptyUsername(t *testing.T) {
	repo := &MockUserRepository{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := newTestUserService(repo)

	user, err := service.Create(context.Background(), commands.CreateUserCommand{
		Username: "",
		Password: "password123",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected user to be nil on error")
	}
	if repo.CreatedUser != nil {
		t.Fatal("repository should not be called")
	}
}

func TestCreateUser_ShortPassword(t *testing.T) {
	repo := &MockUserRepository{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := newTestUserService(repo)

	user, err := service.Create(context.Background(), commands.CreateUserCommand{
		Username: "testuser",
		Password: "123",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected user to be nil on error")
	}
	if repo.CreatedUser != nil {
		t.Fatal("repository should not be called")
	}
}
