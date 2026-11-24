package commands

import "github.com/google/uuid"

type CreateCardCommand struct {
	// UserID is the authenticated user. It must come from the request context,
	// never from client input.
	UserID uuid.UUID
	DeckID uuid.UUID
	Front  string
	Back   string
}
