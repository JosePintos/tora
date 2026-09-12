package ports

import "github.com/google/uuid"

// Why a port? Because UserService should know:
// “I need something that generates authentication tokens.”
// It should not know about JWT specifically.
// Later the adapter will implement that with JWT.

type TokenProvider interface {
	GenerateToken(userID uuid.UUID) (string, error)
	ParseToken(token string) (uuid.UUID, error)
}
