// Package deckaccess implements the card context's DeckAccess port by asking the
// deck context. It is the only place where the card context touches deck code,
// and it translates the deck context's errors into the card context's own
// (an anti-corruption layer).
package deckaccess

import (
	"context"
	"errors"

	"github.com/JosePintos/tora/apps/api/internal/card/application/ports"
	deckservices "github.com/JosePintos/tora/apps/api/internal/deck/application/services"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/google/uuid"
)

// DeckOwnership is the part of the deck service this adapter uses. Depending on
// a one-method interface instead of *deckservices.DeckService keeps the adapter
// easy to test.
type DeckOwnership interface {
	GetOwned(ctx context.Context, ownerID, deckID uuid.UUID) (*domain.Deck, error)
	GetOwnedMany(ctx context.Context, ownerID uuid.UUID, deckIDs []uuid.UUID) ([]*domain.Deck, error)
}

type Adapter struct {
	decks DeckOwnership
}

func New(decks DeckOwnership) *Adapter {
	return &Adapter{decks: decks}
}

func (a *Adapter) EnsureOwnsDeck(ctx context.Context, userID, deckID uuid.UUID) error {
	_, err := a.decks.GetOwned(ctx, userID, deckID)
	if errors.Is(err, deckservices.ErrDeckNotFound) {
		return ports.ErrDeckNotFound
	}
	return err
}

func (a *Adapter) EnsureOwnsDecks(ctx context.Context, userID uuid.UUID, deckIDs []uuid.UUID) error {
	_, err := a.decks.GetOwnedMany(ctx, userID, deckIDs)
	if errors.Is(err, deckservices.ErrDeckNotFound) {
		return ports.ErrDeckNotFound
	}
	return err
}

var _ ports.DeckAccess = (*Adapter)(nil)
