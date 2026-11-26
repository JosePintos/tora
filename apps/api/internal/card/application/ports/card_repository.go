package ports

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/google/uuid"
)

type CardRepository interface {
	Create(ctx context.Context, card *domain.Card) error
	ListByDeck(ctx context.Context, deckID uuid.UUID) ([]*domain.Card, error)
	// ListByDecks returns the cards of all the given decks in one query,
	// ordered like ListByDeck within each deck.
	ListByDecks(ctx context.Context, deckIDs []uuid.UUID) ([]*domain.Card, error)
}
