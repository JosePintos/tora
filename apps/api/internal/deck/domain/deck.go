package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyDeckName   = errors.New("deck name cannot be empty")
	ErrDeckNameTooLong = errors.New("deck name cannot exceed 100 characters")
)

type Deck struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	CreatedAt time.Time
}

func NewDeck(ownerID uuid.UUID, name string) (*Deck, error) {
	if err := validateDeckName(name); err != nil {
		return nil, err
	}

	return &Deck{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Name:      name,
		CreatedAt: time.Now(),
	}, nil
}

func (d *Deck) Rename(newName string) error {
	if err := validateDeckName(newName); err != nil {
		return err
	}
	d.Name = newName
	return nil
}

func validateDeckName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyDeckName
	}

	if len(name) > 100 {
		return ErrDeckNameTooLong
	}
	return nil
}
