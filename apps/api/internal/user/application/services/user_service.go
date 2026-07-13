package services

import (
	"context"
	"errors"
	"strings"

	"github.com/JosePintos/tora/apps/api/internal/user/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
)

type UserService struct {
	userRepository ports.UserRepository
	passwordHasher ports.PasswordHasher
}

func NewUserService(userRepository ports.UserRepository, passwordHasher ports.PasswordHasher) *UserService {
	return &UserService{
		userRepository: userRepository,
		passwordHasher: passwordHasher,
	}
}

func (s *UserService) Create(ctx context.Context, cmd commands.CreateUserCommand) (*domain.User, error) {
	if strings.TrimSpace(cmd.Username) == "" {
		return nil, errors.New("username cannot be empty")
	}

	if len(cmd.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	hash, err := s.passwordHasher.Hash(cmd.Password)
	if err != nil {
		return nil, err
	}

	user := domain.NewUser(cmd.Username, hash)

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
