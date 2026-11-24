package bootstrap

import (
	cardservices "github.com/JosePintos/tora/apps/api/internal/card/application/services"
	"github.com/JosePintos/tora/apps/api/internal/card/infrastructure/deckaccess"
	cardpostgres "github.com/JosePintos/tora/apps/api/internal/card/infrastructure/persistence/postgres"
	deckservices "github.com/JosePintos/tora/apps/api/internal/deck/application/services"
	deckpostgres "github.com/JosePintos/tora/apps/api/internal/deck/infrastructure/persistence/postgres"
	"github.com/JosePintos/tora/apps/api/internal/platform/config"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
	userservices "github.com/JosePintos/tora/apps/api/internal/user/application/services"
	"github.com/JosePintos/tora/apps/api/internal/user/infrastructure/auth"
	userpostgres "github.com/JosePintos/tora/apps/api/internal/user/infrastructure/persistence/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	UserService   *userservices.UserService
	DeckService   *deckservices.DeckService
	CardService   *cardservices.CardService
	TokenProvider ports.TokenProvider
}

// the bootstrap needs the config because it is responsible for wiring concrete infrastructure, and the JWT adapter needs the secret from configuration.
func New(db *pgxpool.Pool, cfg *config.Config) *Container {
	queries := sqlc.New(db)

	userRepository := userpostgres.New(queries)

	passwordHasher := auth.NewBcryptHasher()

	tokenProvider := auth.NewJWTProvider(cfg.JWT.Secret)

	userService := userservices.NewUserService(userRepository, passwordHasher, tokenProvider)

	deckRepository := deckpostgres.New(queries)
	deckService := deckservices.NewDeckService(deckRepository)

	// The card context reaches decks only through its DeckAccess port.
	cardRepository := cardpostgres.New(queries)
	cardService := cardservices.NewCardService(cardRepository, deckaccess.New(deckService))

	return &Container{
		UserService:   userService,
		DeckService:   deckService,
		CardService:   cardService,
		TokenProvider: tokenProvider,
	}
}
