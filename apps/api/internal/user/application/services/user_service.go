package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JosePintos/tora/apps/api/internal/user/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
)

var (
	ErrUserPasswordTooShort = errors.New("user password is too short")
	ErrEmptyUsername        = errors.New("username cannot be empty")
	ErrUsernameTooLong      = errors.New("username cannot exceed 100 characters")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPassword      = errors.New("invalid password")
)

type UserService struct {
	userRepository ports.UserRepository
	passwordHasher ports.PasswordHasher
	tokenProvider  ports.TokenProvider
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
	fmt.Printf("User found: %s\n for username: %s\n", user.Username, username)
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

	if err := s.passwordHasher.Compare(user.PasswordHash, cmd.Password); err != nil {
		return "", ErrInvalidPassword
	}

	token, err := s.tokenProvider.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func validateUserCredentials(username string, password string) error {
	if strings.TrimSpace(username) == "" {
		return ErrEmptyUsername
	}

	if len(username) > 100 {
		return ErrUsernameTooLong
	}
	if len(password) < 8 {
		return ErrUserPasswordTooShort
	}
	return nil
}
