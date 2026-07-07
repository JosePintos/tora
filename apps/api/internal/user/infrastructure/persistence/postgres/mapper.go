package postgres

import (
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

func toDomain(user sqlc.User) *domain.User {
	return &domain.User{
		ID:        user.ID.Bytes,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Time,
	}
}

func toDB(user *domain.User) sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: user.ID, Valid: true},
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: user.CreatedAt, Valid: true},
	}
}
