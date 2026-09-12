package ports

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/google/uuid"
)

type DeckRepository interface {
	Create(ctx context.Context, deck *domain.Deck) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Deck, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Deck, error)
	Update(ctx context.Context, deck *domain.Deck) error
	Delete(ctx context.Context, id uuid.UUID) error
}
