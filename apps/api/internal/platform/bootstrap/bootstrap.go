package bootsrap

import (
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/JosePintos/tora/apps/api/internal/user/application/services"
	"github.com/JosePintos/tora/apps/api/internal/user/infrastructure/auth"
	"github.com/JosePintos/tora/apps/api/internal/user/infrastructure/persistence/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	UserService *services.UserService
}

func New(db *pgxpool.Pool) *Container {
	queries := sqlc.New(db)

	userRepository := postgres.New(queries)

	passwordHasher := auth.NewBcryptHasher()

	userService := services.NewUserService(userRepository, passwordHasher)

	return &Container{
		UserService: userService,
	}
}
