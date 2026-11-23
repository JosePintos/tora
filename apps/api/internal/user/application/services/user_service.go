package services

import (
	"context"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/JosePintos/tora/apps/api/internal/user/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
)

var (
	ErrUserPasswordTooShort = apperror.New(apperror.Invalid, "user password is too short")
	ErrEmptyUsername        = domain.ErrEmptyUsername // alias: the domain owns this rule
	ErrUsernameTooLong      = apperror.New(apperror.Invalid, "username cannot exceed 100 characters")
	ErrUserNotFound         = apperror.New(apperror.NotFound, "user not found")

	// ErrInvalidCredentials is returned by Login for both an unknown username and
	// a wrong password. Using one error (and one response time) for both stops
	// callers from discovering which usernames exist (user enumeration).
	ErrInvalidCredentials = apperror.New(apperror.Unauthenticated, "invalid username or password")
)

// dummyPassword is hashed once to give Login a real hash to compare against when
// the username does not exist. Its value is irrelevant; it never matches anything
// that grants access.
const dummyPassword = "tora-login-timing-equalizer"

type UserService struct {
	userRepository ports.UserRepository
	passwordHasher ports.PasswordHasher
	tokenProvider  ports.TokenProvider

	dummyHashOnce sync.Once
	dummyHash     string
}

func NewUserService(userRepository ports.UserRepository, passwordHasher ports.PasswordHasher, tokenProvider ports.TokenProvider) *UserService {
	return &UserService{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
		tokenProvider:  tokenProvider,
	}
}

func (s *UserService) Create(ctx context.Context, cmd commands.CreateUserCommand) (*domain.User, error) {
	if err := validateUserCredentials(cmd.Username, cmd.Password); err != nil {
		return nil, err
	}

	hash, err := s.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(cmd.Username, hash)
	if err != nil {
		return nil, err
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepository.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, cmd commands.LoginCommand) (string, error) {
	user, err := s.userRepository.GetByUsername(ctx, cmd.Username)
	if err != nil {
		return "", err
	}
	if user == nil {
		// Do the same expensive work as a real login so an unknown username
		// doesn't answer measurably faster than a wrong password.
		s.compareWithDummyHash(cmd.Password)
		return "", ErrInvalidCredentials
	}

	if err := s.passwordHasher.Compare(user.PasswordHash, cmd.Password); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.tokenProvider.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// compareWithDummyHash runs a password comparison whose result is ignored. The
// dummy hash is created lazily through the PasswordHasher port, so it uses the
// same algorithm and cost as real password hashes without this service knowing
// which algorithm that is.
func (s *UserService) compareWithDummyHash(password string) {
	s.dummyHashOnce.Do(func() {
		// Hashing a short constant can't fail with bcrypt. If a future hasher
		// did fail, the comparison below fails fast: login stays correct, only
		// the timing protection is lost.
		s.dummyHash, _ = s.passwordHasher.Hash(dummyPassword)
	})
	_ = s.passwordHasher.Compare(s.dummyHash, password)
}

func validateUserCredentials(username string, password string) error {
	if strings.TrimSpace(username) == "" {
		return ErrEmptyUsername
	}

	if utf8.RuneCountInString(username) > 100 {
		return ErrUsernameTooLong
	}
	if len(password) < 8 {
		return ErrUserPasswordTooShort
	}
	return nil
}
