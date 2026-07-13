package postgres

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
	"github.com/google/uuid"
)

type Repository struct {
	queries *sqlc.Queries
}

func New(queries *sqlc.Queries) *Repository {
	return &Repository{
		queries: queries,
	}
}

func (r *Repository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.queries.CreateUser(ctx, toDB(user))

	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	panic("not implemented")
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return toDomain(user), nil
}

var _ ports.UserRepository = (*Repository)(nil)
