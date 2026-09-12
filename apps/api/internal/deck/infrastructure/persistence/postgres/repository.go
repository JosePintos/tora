package postgres

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/deck/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
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

func (r *Repository) Create(ctx context.Context, deck *domain.Deck) error {
	_, err := r.queries.CreateDeck(ctx, toDB(deck))
	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deck, error) {
	deck, err := r.queries.GetDeckByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, err
	}
	return toDomain(deck), nil
}

func (r *Repository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Deck, error) {
	decks, err := r.queries.ListDecksByOwner(ctx, pgtype.UUID{Bytes: ownerID, Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Deck, 0, len(decks))

	for _, deck := range decks {
		result = append(result, toDomain(deck))
	}

	return result, nil
}

func (r *Repository) Update(ctx context.Context, deck *domain.Deck) error {
	panic("not implemented")
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("not implemented")
}

var _ ports.DeckRepository = (*Repository)(nil)
