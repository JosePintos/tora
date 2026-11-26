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

// GetOwnedMany is the batch version of GetOwned, for callers that need many
// decks at once (for example a DataLoader). It returns the decks only if ownerID
// owns every one of them; if any is missing or owned by someone else it returns
// ErrDeckNotFound, without saying which. Duplicate IDs are allowed.
func (s *DeckService) GetOwnedMany(ctx context.Context, ownerID uuid.UUID, deckIDs []uuid.UUID) ([]*domain.Deck, error) {
	if len(deckIDs) == 0 {
		return []*domain.Deck{}, nil
	}

	decks, err := s.deckRepository.ListByIDs(ctx, deckIDs)
	if err != nil {
		return nil, err
	}

	owned := make(map[uuid.UUID]bool, len(decks))
	for _, deck := range decks {
		if !deck.IsOwnedBy(ownerID) {
			return nil, ErrDeckNotFound
		}
		owned[deck.ID] = true
	}
	for _, id := range deckIDs {
		if !owned[id] {
			return nil, ErrDeckNotFound
		}
	}

	return decks, nil
}

func (s *DeckService) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Deck, error) {
	return s.deckRepository.ListByOwner(ctx, ownerID)
}
