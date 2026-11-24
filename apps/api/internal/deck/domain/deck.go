package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/google/uuid"
)

var (
	ErrEmptyDeckName   = apperror.New(apperror.Invalid, "deck name cannot be empty")
	ErrDeckNameTooLong = apperror.New(apperror.Invalid, "deck name cannot exceed 100 characters")
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

// IsOwnedBy reports whether userID owns the deck. Only the owner may read or
// change a deck and its cards.
func (d *Deck) IsOwnedBy(userID uuid.UUID) bool {
	return d.OwnerID == userID
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

	// Count characters, not bytes: a Japanese character is 3 bytes in UTF-8.
	if utf8.RuneCountInString(name) > 100 {
		return ErrDeckNameTooLong
	}
	return nil
}
