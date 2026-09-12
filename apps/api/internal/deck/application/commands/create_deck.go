package commands

import "github.com/google/uuid"

type CreateDeckCommand struct {
	OwnerID uuid.UUID
	Name    string
}
