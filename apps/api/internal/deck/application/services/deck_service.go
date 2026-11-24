package services

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/JosePintos/tora/apps/api/internal/deck/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/deck/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/google/uuid"
)

// ErrDeckNotFound is returned when a deck doesn't exist or belongs to another
// user. Both cases get the same error so callers can't probe which deck IDs
// exist.
var ErrDeckNotFound = apperror.New(apperror.NotFound, "deck not found")

type DeckService struct {
	deckRepository ports.DeckRepository
}

func NewDeckService(deckRepository ports.DeckRepository) *DeckService {
	return &DeckService{
		deckRepository: deckRepository,
	}
}

func (s *DeckService) Create(ctx context.Context, cmd commands.CreateDeckCommand) (*domain.Deck, error) {
	deck, err := domain.NewDeck(cmd.OwnerID, cmd.Name)
	if err != nil {
		return nil, err
	}

	if err := s.deckRepository.Create(ctx, deck); err != nil {
		return nil, err
	}

	return deck, nil
}

// GetOwned returns the deck with deckID if ownerID owns it, or ErrDeckNotFound.
func (s *DeckService) GetOwned(ctx context.Context, ownerID, deckID uuid.UUID) (*domain.Deck, error) {
	deck, err := s.deckRepository.GetByID(ctx, deckID)
	if err != nil {
		return nil, err
	}
	if deck == nil || !deck.IsOwnedBy(ownerID) {
		return nil, ErrDeckNotFound
	}
	return deck, nil
}

func (s *DeckService) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Deck, error) {
	return s.deckRepository.ListByOwner(ctx, ownerID)
}
