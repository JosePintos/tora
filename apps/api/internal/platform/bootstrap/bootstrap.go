package bootsrap

import (
	deckservices "github.com/JosePintos/tora/apps/api/internal/deck/application/services"
	deckpostgres "github.com/JosePintos/tora/apps/api/internal/deck/infrastructure/persistence/postgres"
	"github.com/JosePintos/tora/apps/api/internal/platform/config"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/JosePintos/tora/apps/api/internal/user/application/services"
	userservices "github.com/JosePintos/tora/apps/api/internal/user/application/services"
	"github.com/JosePintos/tora/apps/api/internal/user/infrastructure/auth"
	userpostgres "github.com/JosePintos/tora/apps/api/internal/user/infrastructure/persistence/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	UserService *userservices.UserService
	DeckService *deckservices.DeckService
}

// the bootstrap needs the config because it is responsible for wiring concrete infrastructure, and the JWT adapter needs the secret from configuration.
func New(db *pgxpool.Pool, cfg *config.Config) *Container {
	queries := sqlc.New(db)

	userRepository := userpostgres.New(queries)

	passwordHasher := auth.NewBcryptHasher()

	tokenProvider := auth.NewJWTProvider(cfg.JWT.Secret)

	userService := services.NewUserService(userRepository, passwordHasher, tokenProvider)

	deckRepository := deckpostgres.New(queries)
	deckService := deckservices.NewDeckService(deckRepository)

	return &Container{
		UserService: userService,
		DeckService: deckService,
	}
}
