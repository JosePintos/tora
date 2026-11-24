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
	CreateFn           func(context.Context, *domain.User) error
	CreatedUser        *domain.User
	CreateCalls        int
	GetByUsernameFn    func(context.Context, string) (*domain.User, error)
	GetByUsernameCalls int
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
	m.GetByUsernameCalls++
	if m.GetByUsernameFn != nil {
		return m.GetByUsernameFn(ctx, username)
	}
	return nil, nil
}

type MockPasswordHasher struct {
	HashFn    func(string) (string, error)
	CompareFn func(hash, password string) error
}

func (m MockPasswordHasher) Hash(password string) (string, error) {
	if m.HashFn != nil {
		return m.HashFn(password)
	}
	return "", nil
}

func (m MockPasswordHasher) Compare(hash, password string) error {
	if m.CompareFn != nil {
		return m.CompareFn(hash, password)
	}
	return nil
}

type MockTokenProvider struct {
	GenerateTokenFn func(uuid.UUID) (string, error)
	ParseTokenFn    func(string) (uuid.UUID, error)
}

func (m *MockTokenProvider) GenerateToken(userID uuid.UUID) (string, error) {
	if m.GenerateTokenFn != nil {
		return m.GenerateTokenFn(userID)
	}
	return "", nil
}

func (m *MockTokenProvider) ParseToken(token string) (uuid.UUID, error) {
	if m.ParseTokenFn != nil {
		return m.ParseTokenFn(token)
	}
	return uuid.Nil, nil
}

func newTestUserService(repo ports.UserRepository) *UserService {
	hasher := MockPasswordHasher{
		HashFn: func(password string) (string, error) {
			return "hashed_password", nil
		},
	}

	return NewUserService(repo, hasher, &MockTokenProvider{})
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

	service := NewUserService(repo, hasher, &MockTokenProvider{})
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
	if !errors.Is(err, ErrEmptyUsername) {
		t.Fatalf("expected ErrEmptyUsername, got %v", err)
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
	if !errors.Is(err, ErrUserPasswordTooShort) {
		t.Fatalf("expected ErrUserPasswordTooShort, got %v", err)
	}
	if user != nil {
		t.Fatal("expected user to be nil on error")
	}
	if repo.CreatedUser != nil {
		t.Fatal("repository should not be called")
	}
}

func TestCreateUser_UsernameTooLong(t *testing.T) {
	repo := &MockUserRepository{
		CreateFn: func(ctx context.Context, user *domain.User) error {
			return nil
		},
	}

	service := newTestUserService(repo)

	longUsername := "This is a very long deck name that exceeds the maximum allowed length of one hundred characters. It should trigger an error."
	user, err := service.Create(context.Background(), commands.CreateUserCommand{
		Username: longUsername,
		Password: "password123",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUsernameTooLong) {
		t.Fatalf("expected ErrUsernameTooLong, got %v", err)
	}
	if user != nil {
		t.Fatal("expected user to be nil on error")
	}
	if repo.CreatedUser != nil {
		t.Fatal("repository should not be called")
	}
}

func TestLogin_Successful(t *testing.T) {
	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			return &domain.User{
				ID:           uuid.New(),
				Username:     "testuser",
				PasswordHash: "hashedpassword",
			}, nil
		},
	}

	hasher := &MockPasswordHasher{
		CompareFn: func(hash, password string) error {
			return nil
		},
	}

	tokenProvider := &MockTokenProvider{
		GenerateTokenFn: func(userID uuid.UUID) (string, error) {
			return "token", nil
		},
	}

	service := NewUserService(repo, hasher, tokenProvider)

	token, err := service.Login(context.Background(), commands.LoginCommand{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "token" {
		t.Fatalf("expected token to be 'token', got %v", token)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			return &domain.User{
				ID:           uuid.New(),
				Username:     "testuser",
				PasswordHash: "hashedpassword",
			}, nil
		},
	}

	hasher := &MockPasswordHasher{
		CompareFn: func(hash, password string) error {
			return errors.New("bcrypt: hashedPassword is not the hash of the given password")
		},
	}

	tokenProvider := &MockTokenProvider{
		GenerateTokenFn: func(userID uuid.UUID) (string, error) {
			return "token", nil
		},
	}

	service := NewUserService(repo, hasher, tokenProvider)

	_, err := service.Login(context.Background(), commands.LoginCommand{
		Username: "testuser",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_FailGenerateToken(t *testing.T) {
	var errGenerateToken = errors.New("failed to generate token")
	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			return &domain.User{
				ID:           uuid.New(),
				Username:     "testuser",
				PasswordHash: "hashedpassword",
			}, nil
		},
	}

	hasher := &MockPasswordHasher{
		CompareFn: func(hash, password string) error {
			return nil
		},
	}

	tokenProvider := &MockTokenProvider{
		GenerateTokenFn: func(userID uuid.UUID) (string, error) {
			return "", errGenerateToken
		},
	}

	service := NewUserService(repo, hasher, tokenProvider)

	_, err := service.Login(context.Background(), commands.LoginCommand{
		Username: "testuser",
		Password: "password123",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, errGenerateToken) {
		t.Fatalf("expected errGenerateToken, got %v", err)
	}
}

func TestLogin_UnknownUsername(t *testing.T) {
	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			return nil, nil // repositories return (nil, nil) when no row matches
		},
	}

	var comparedHashes []string
	hasher := &MockPasswordHasher{
		HashFn: func(password string) (string, error) {
			return "dummy-hash", nil
		},
		CompareFn: func(hash, password string) error {
			comparedHashes = append(comparedHashes, hash)
			return errors.New("mismatch")
		},
	}

	tokenProvider := &MockTokenProvider{
		GenerateTokenFn: func(userID uuid.UUID) (string, error) {
			t.Fatal("no token must be generated for an unknown user")
			return "", nil
		},
	}

	service := NewUserService(repo, hasher, tokenProvider)

	token, err := service.Login(context.Background(), commands.LoginCommand{
		Username: "ghost",
		Password: "password123",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if token != "" {
		t.Fatal("expected token to be empty on error")
	}
	if len(comparedHashes) != 1 || comparedHashes[0] != "dummy-hash" {
		t.Fatalf("expected one comparison against the dummy hash (timing equalization), got %v", comparedHashes)
	}
}

func TestLogin_UnknownUserAndWrongPasswordAreIndistinguishable(t *testing.T) {
	existing := &domain.User{ID: uuid.New(), Username: "alice", PasswordHash: "real-hash"}

	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			if username == "alice" {
				return existing, nil
			}
			return nil, nil
		},
	}
	hasher := &MockPasswordHasher{
		CompareFn: func(hash, password string) error { return errors.New("mismatch") },
	}

	service := NewUserService(repo, hasher, &MockTokenProvider{})

	_, unknownErr := service.Login(context.Background(), commands.LoginCommand{Username: "nobody", Password: "x"})
	_, wrongPasswordErr := service.Login(context.Background(), commands.LoginCommand{Username: "alice", Password: "x"})

	if !errors.Is(unknownErr, ErrInvalidCredentials) || !errors.Is(wrongPasswordErr, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for both, got %q and %q", unknownErr, wrongPasswordErr)
	}
	if unknownErr.Error() != wrongPasswordErr.Error() {
		t.Fatalf("expected identical messages, got %q and %q", unknownErr, wrongPasswordErr)
	}
}

func TestLogin_DummyHashIsComputedOnce(t *testing.T) {
	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			return nil, nil
		},
	}

	hashCalls := 0
	hasher := &MockPasswordHasher{
		HashFn: func(password string) (string, error) {
			hashCalls++
			return "dummy-hash", nil
		},
		CompareFn: func(hash, password string) error { return errors.New("mismatch") },
	}

	service := NewUserService(repo, hasher, &MockTokenProvider{})

	for range 3 {
		_, _ = service.Login(context.Background(), commands.LoginCommand{Username: "ghost", Password: "x"})
	}

	if hashCalls != 1 {
		t.Fatalf("expected the dummy hash to be created once, got %d Hash calls", hashCalls)
	}
}

func TestLogin_RepositoryFailureIsNotReportedAsBadCredentials(t *testing.T) {
	errDatabase := errors.New("connection refused")
	repo := &MockUserRepository{
		GetByUsernameFn: func(ctx context.Context, username string) (*domain.User, error) {
			return nil, errDatabase
		},
	}

	service := NewUserService(repo, &MockPasswordHasher{}, &MockTokenProvider{})

	_, err := service.Login(context.Background(), commands.LoginCommand{Username: "alice", Password: "x"})
	if !errors.Is(err, errDatabase) {
		t.Fatalf("expected the database error to pass through, got %v", err)
	}
}
