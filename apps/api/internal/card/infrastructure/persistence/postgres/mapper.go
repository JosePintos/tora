package postgres

import (
	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func toDomain(card sqlc.Card) *domain.Card {
	return &domain.Card{
		ID:        card.ID.Bytes,
		DeckID:    card.DeckID.Bytes,
		Front:     card.Front,
		Back:      card.Back,
		CreatedAt: card.CreatedAt.Time,
	}
}

func toDB(card *domain.Card) sqlc.CreateCardParams {
	return sqlc.CreateCardParams{
		ID:        pgtype.UUID{Bytes: card.ID, Valid: true},
		DeckID:    pgtype.UUID{Bytes: card.DeckID, Valid: true},
		Front:     card.Front,
		Back:      card.Back,
		CreatedAt: pgtype.Timestamptz{Time: card.CreatedAt, Valid: true},
	}
}
