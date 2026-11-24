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
