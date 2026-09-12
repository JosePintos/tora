package postgres

import (
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func toDomain(deck sqlc.Deck) *domain.Deck {
	return &domain.Deck{
		ID:        deck.ID.Bytes,
		OwnerID:   deck.OwnerID.Bytes,
		Name:      deck.Name,
		CreatedAt: deck.CreatedAt.Time,
	}
}

func toDB(deck *domain.Deck) sqlc.CreateDeckParams {
	return sqlc.CreateDeckParams{
		ID:        pgtype.UUID{Bytes: deck.ID, Valid: true},
		OwnerID:   pgtype.UUID{Bytes: deck.OwnerID, Valid: true},
		Name:      deck.Name,
		CreatedAt: pgtype.Timestamptz{Time: deck.CreatedAt, Valid: true},
	}
}
