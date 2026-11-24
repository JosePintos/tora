package domain

import (
	"strings"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/google/uuid"
)

var (
	ErrEmptyCardFront = apperror.New(apperror.Invalid, "card front cannot be empty")
	ErrEmptyCardBack  = apperror.New(apperror.Invalid, "card back cannot be empty")
)

type Card struct {
	ID        uuid.UUID
	DeckID    uuid.UUID
	Front     string
	Back      string
	CreatedAt time.Time
}

func NewCard(deckID uuid.UUID, front, back string) (*Card, error) {
	front = strings.TrimSpace(front)
	back = strings.TrimSpace(back)

	if front == "" {
		return nil, ErrEmptyCardFront
	}
	if back == "" {
		return nil, ErrEmptyCardBack
	}
	return &Card{
		ID:        uuid.New(),
		DeckID:    deckID,
		Front:     front,
		Back:      back,
		CreatedAt: time.Now(),
	}, nil
}
