package domain

import (
	"strings"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/google/uuid"
)

var (
	ErrEmptyUsername = apperror.New(apperror.Invalid, "username cannot be empty")
	// ErrUsernameTaken is returned by repositories when the unique username rule is broken.
	ErrUsernameTaken = apperror.New(apperror.Conflict, "username is already taken")
)

// No json tags porque no es una API model
// Es business model
// JSON son para graphql/http
// The domain should be pure Go.
type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Behavior belongs in the domain.
// Not in services.
func (u User) IsRegistered() bool {
	return u.ID != uuid.Nil
}

func NewUser(username, passwordHash string) (*User, error) {

	if strings.TrimSpace(username) == "" {
		return nil, ErrEmptyUsername
	}

	return &User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}, nil
}
