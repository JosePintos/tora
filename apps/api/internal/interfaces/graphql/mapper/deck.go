package mapper

import (
	"time"

	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/model"
)

func ToGraphQLDeck(deck *domain.Deck) *model.Deck {
	return &model.Deck{
		ID:        deck.ID.String(),
		Name:      deck.Name,
		OwnerID:   deck.OwnerID.String(),
		CreatedAt: deck.CreatedAt.Format(time.RFC3339),
	}
}
