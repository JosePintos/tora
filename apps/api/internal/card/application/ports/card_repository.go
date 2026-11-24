package ports

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/google/uuid"
)

type CardRepository interface {
	Create(ctx context.Context, card *domain.Card) error
	ListByDeck(ctx context.Context, deckID uuid.UUID) ([]*domain.Card, error)
}
