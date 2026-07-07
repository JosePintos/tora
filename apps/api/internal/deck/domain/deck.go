package domain

import (
	"time"

	"github.com/google/uuid"
)

type Deck struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	CreatedAt time.Time
}

func NewDeck(ownerID uuid.UUID, name string) *Deck {
	return &Deck{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		Name:      name,
		CreatedAt: time.Now(),
	}
}
