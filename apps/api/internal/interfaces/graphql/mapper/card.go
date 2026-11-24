package mapper

import (
	"time"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/model"
)

func ToGraphQLCard(card *domain.Card) *model.Card {
	return &model.Card{
		ID:        card.ID.String(),
		DeckID:    card.DeckID.String(),
		Front:     card.Front,
		Back:      card.Back,
		CreatedAt: card.CreatedAt.Format(time.RFC3339),
	}
}

func ToGraphQLCards(cards []*domain.Card) []*model.Card {
	result := make([]*model.Card, 0, len(cards))
	for _, card := range cards {
		result = append(result, ToGraphQLCard(card))
	}
	return result
}
