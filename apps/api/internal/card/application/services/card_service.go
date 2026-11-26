package services

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/card/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/card/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/google/uuid"
)

type CardService struct {
	cardRepository ports.CardRepository
	deckAccess     ports.DeckAccess
}

func NewCardService(cardRepository ports.CardRepository, deckAccess ports.DeckAccess) *CardService {
	return &CardService{
		cardRepository: cardRepository,
		deckAccess:     deckAccess,
	}
}

// Create adds a card to a deck the user owns.
func (s *CardService) Create(ctx context.Context, cmd commands.CreateCardCommand) (*domain.Card, error) {
	card, err := domain.NewCard(cmd.DeckID, cmd.Front, cmd.Back)
	if err != nil {
		return nil, err
	}

	if err := s.deckAccess.EnsureOwnsDeck(ctx, cmd.UserID, cmd.DeckID); err != nil {
		return nil, err
	}

	if err := s.cardRepository.Create(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

// ListByDeck returns the cards of a deck the user owns, oldest first.
func (s *CardService) ListByDeck(ctx context.Context, userID, deckID uuid.UUID) ([]*domain.Card, error) {
	if err := s.deckAccess.EnsureOwnsDeck(ctx, userID, deckID); err != nil {
		return nil, err
	}

	return s.cardRepository.ListByDeck(ctx, deckID)
}

// ListByDecks returns the cards of several decks the user owns, keyed by deck
// ID, using one ownership check and one query whatever the number of decks.
// Every requested deck has an entry; a deck without cards maps to an empty
// slice. If any deck is missing or not owned, the whole call fails with
// ports.ErrDeckNotFound.
func (s *CardService) ListByDecks(ctx context.Context, userID uuid.UUID, deckIDs []uuid.UUID) (map[uuid.UUID][]*domain.Card, error) {
	result := make(map[uuid.UUID][]*domain.Card, len(deckIDs))
	if len(deckIDs) == 0 {
		return result, nil
	}

	if err := s.deckAccess.EnsureOwnsDecks(ctx, userID, deckIDs); err != nil {
		return nil, err
	}

	cards, err := s.cardRepository.ListByDecks(ctx, deckIDs)
	if err != nil {
		return nil, err
	}

	for _, id := range deckIDs {
		result[id] = []*domain.Card{}
	}
	for _, card := range cards {
		result[card.DeckID] = append(result[card.DeckID], card)
	}

	return result, nil
}
