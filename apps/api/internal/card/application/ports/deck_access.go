package ports

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/google/uuid"
)

// ErrDeckNotFound is returned when the deck doesn't exist or isn't owned by the
// user. Both cases look the same so deck IDs of other users can't be probed.
var ErrDeckNotFound = apperror.New(apperror.NotFound, "deck not found")

// DeckAccess is what the card context needs to know about decks. Cards belong
// to decks, but deck ownership is the deck context's business, so the card
// context asks through this port instead of depending on the deck packages.
type DeckAccess interface {
	// EnsureOwnsDeck returns nil if userID owns deckID, ErrDeckNotFound if the
	// deck is missing or owned by someone else, or another error on failure.
	EnsureOwnsDeck(ctx context.Context, userID, deckID uuid.UUID) error

	// EnsureOwnsDecks is the batch version of EnsureOwnsDeck: one check for many
	// decks, so batch loading doesn't need one ownership query per deck. It
	// returns ErrDeckNotFound if any deck is missing or owned by someone else.
	EnsureOwnsDecks(ctx context.Context, userID uuid.UUID, deckIDs []uuid.UUID) error
}
