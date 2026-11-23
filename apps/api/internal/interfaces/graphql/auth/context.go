package auth

import (
	"context"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/google/uuid"
)

// ErrUnauthenticated is returned by resolvers that require a logged-in user.
var ErrUnauthenticated = apperror.New(apperror.Unauthenticated, "unauthenticated")

type contextKey string

const userIDKey contextKey = "userID"

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
