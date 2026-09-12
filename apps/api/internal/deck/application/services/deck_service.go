package services

import (
	"context"
	"errors"
	"strings"

	"github.com/JosePintos/tora/apps/api/internal/deck/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/deck/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/google/uuid"
)

type DeckService struct {
	deckRepository ports.DeckRepository
}

func NewDeckService(deckRepository ports.DeckRepository) *DeckService {
	return &DeckService{
		deckRepository: deckRepository,
	}
}

func (s *DeckService) Create(ctx context.Context, cmd commands.CreateDeckCommand) (*domain.Deck, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return nil, errors.New("deckname cannot be empty")
	}

	deck, err := domain.NewDeck(cmd.OwnerID, cmd.Name)
	if err != nil {
		return nil, err
	}

	if err := s.deckRepository.Create(ctx, deck); err != nil {
		return nil, err
	}

	return deck, nil
}

func (s *DeckService) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Deck, error) {
	return s.deckRepository.ListByOwner(ctx, ownerID)
}
