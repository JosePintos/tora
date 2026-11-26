package postgres

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/card/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Repository struct {
	queries *sqlc.Queries
}

func New(queries *sqlc.Queries) *Repository {
	return &Repository{
		queries: queries,
	}
}

func (r *Repository) Create(ctx context.Context, card *domain.Card) error {
	_, err := r.queries.CreateCard(ctx, toDB(card))
	return err
}

func (r *Repository) ListByDeck(ctx context.Context, deckID uuid.UUID) ([]*domain.Card, error) {
	cards, err := r.queries.ListCardsByDeck(ctx, pgtype.UUID{Bytes: deckID, Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Card, 0, len(cards))
	for _, card := range cards {
		result = append(result, toDomain(card))
	}

	return result, nil
}

func (r *Repository) ListByDecks(ctx context.Context, deckIDs []uuid.UUID) ([]*domain.Card, error) {
	pgIDs := make([]pgtype.UUID, len(deckIDs))
	for i, id := range deckIDs {
		pgIDs[i] = pgtype.UUID{Bytes: id, Valid: true}
	}

	cards, err := r.queries.ListCardsByDecks(ctx, pgIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Card, 0, len(cards))
	for _, card := range cards {
		result = append(result, toDomain(card))
	}

	return result, nil
}

var _ ports.CardRepository = (*Repository)(nil)
