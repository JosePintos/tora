package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
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
		return nil, errors.New("username cannot be empty")
	}

	return &User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}, nil
}
