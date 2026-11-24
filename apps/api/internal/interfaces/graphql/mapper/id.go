package mapper

import (
	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/google/uuid"
)

// ErrInvalidID is returned when a GraphQL ID argument is not a valid UUID.
var ErrInvalidID = apperror.New(apperror.Invalid, "invalid id")

// ParseID converts a GraphQL ID (a string on the wire) into a UUID.
func ParseID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, ErrInvalidID
	}
	return parsed, nil
}
