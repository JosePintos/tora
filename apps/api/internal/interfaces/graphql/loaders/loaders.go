// Package loaders holds the request-scoped DataLoaders used by GraphQL
// resolvers to batch lookups that would otherwise run once per parent object
// (the N+1 query problem).
//
// Loaders cache results, so they must never outlive a request: a shared loader
// would serve one user's data to another. Middleware creates a fresh set for
// every request.
package loaders

import (
	"context"
	"net/http"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/auth"
	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

// CardLister is the part of the card service the loaders need.
type CardLister interface {
	ListByDecks(ctx context.Context, userID uuid.UUID, deckIDs []uuid.UUID) (map[uuid.UUID][]*domain.Card, error)
}

// Loaders are the DataLoaders for one request.
type Loaders struct {
	// CardsByDeck loads a deck's cards; all Load calls made while a batch is
	// open become one CardLister.ListByDecks call.
	CardsByDeck *dataloadgen.Loader[uuid.UUID, []*domain.Card]
}

// New builds the loaders for one request made by userID.
func New(cards CardLister, userID uuid.UUID, options ...dataloadgen.Option) *Loaders {
	return &Loaders{
		CardsByDeck: dataloadgen.NewLoader(cardsByDeck(cards, userID), options...),
	}
}

// cardsByDeck returns the batch function behind CardsByDeck. Its contract with
// dataloadgen: return exactly one value per key, in the same order as keys. A
// single error fails every key in the batch.
func cardsByDeck(cards CardLister, userID uuid.UUID) func(ctx context.Context, deckIDs []uuid.UUID) ([][]*domain.Card, []error) {
	return func(ctx context.Context, deckIDs []uuid.UUID) ([][]*domain.Card, []error) {
		byDeck, err := cards.ListByDecks(ctx, userID, deckIDs)
		if err != nil {
			return nil, []error{err}
		}

		result := make([][]*domain.Card, len(deckIDs))
		for i, id := range deckIDs {
			result[i] = byDeck[id]
		}
		return result, nil
	}
}

type contextKey struct{}

// Middleware gives every authenticated request its own Loaders. It must run
// after auth.AuthMiddleware, because loaders are tied to the requesting user.
// Unauthenticated requests get none; resolvers reject them before loading.
func Middleware(cards CardLister, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, New(cards, userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// For returns the request's loaders, or nil if Middleware didn't run or the
// request is unauthenticated.
func For(ctx context.Context) *Loaders {
	l, _ := ctx.Value(contextKey{}).(*Loaders)
	return l
}
