package loaders

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/auth"
	"github.com/google/uuid"
	"github.com/vikstrous/dataloadgen"
)

// mockCardLister records every batch it receives.
type mockCardLister struct {
	mu      sync.Mutex
	batches [][]uuid.UUID
	userIDs []uuid.UUID
	cards   map[uuid.UUID][]*domain.Card
	err     error
}

func (m *mockCardLister) ListByDecks(_ context.Context, userID uuid.UUID, deckIDs []uuid.UUID) (map[uuid.UUID][]*domain.Card, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batches = append(m.batches, deckIDs)
	m.userIDs = append(m.userIDs, userID)
	return m.cards, m.err
}

func card(t *testing.T, deckID uuid.UUID, front string) *domain.Card {
	t.Helper()
	c, err := domain.NewCard(deckID, front, "back")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	return c
}

// loadConcurrently calls Load once per deck from separate goroutines, the way
// gqlgen runs the Deck.cards resolvers of a list of decks.
func loadConcurrently(l *Loaders, deckIDs []uuid.UUID) ([][]*domain.Card, []error) {
	results := make([][]*domain.Card, len(deckIDs))
	errs := make([]error, len(deckIDs))
	var wg sync.WaitGroup
	for i, id := range deckIDs {
		wg.Go(func() {
			results[i], errs[i] = l.CardsByDeck.Load(context.Background(), id)
		})
	}
	wg.Wait()
	return results, errs
}

func TestCardsByDeck_BatchesConcurrentLoadsIntoOneCall(t *testing.T) {
	userID := uuid.New()
	deckIDs := make([]uuid.UUID, 50)
	cards := map[uuid.UUID][]*domain.Card{}
	for i := range deckIDs {
		deckIDs[i] = uuid.New()
		cards[deckIDs[i]] = []*domain.Card{card(t, deckIDs[i], deckIDs[i].String())}
	}
	lister := &mockCardLister{cards: cards}

	results, errs := loadConcurrently(New(lister, userID), deckIDs)

	if len(lister.batches) != 1 {
		t.Fatalf("expected 50 loads to become 1 batch call, got %d calls", len(lister.batches))
	}
	if len(lister.batches[0]) != 50 || lister.userIDs[0] != userID {
		t.Fatalf("expected one call with all 50 decks for the request's user, got %d decks for %v", len(lister.batches[0]), lister.userIDs[0])
	}
	for i, id := range deckIDs {
		if errs[i] != nil {
			t.Fatalf("deck %d: unexpected error %v", i, errs[i])
		}
		if len(results[i]) != 1 || results[i][0].DeckID != id {
			t.Fatalf("deck %d received another deck's cards: %v", i, results[i])
		}
	}
}

func TestCardsByDeck_DeckWithoutCardsGetsNoCards(t *testing.T) {
	empty := uuid.New()
	lister := &mockCardLister{cards: map[uuid.UUID][]*domain.Card{}}

	got, err := New(lister, uuid.New()).CardsByDeck.Load(context.Background(), empty)
	if err != nil {
		t.Fatalf("a deck without cards must not be an error, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no cards, got %v", got)
	}
}

func TestCardsByDeck_ErrorReachesEveryCallerInTheBatch(t *testing.T) {
	errDatabase := errors.New("connection refused")
	lister := &mockCardLister{err: errDatabase}

	_, errs := loadConcurrently(New(lister, uuid.New()), []uuid.UUID{uuid.New(), uuid.New(), uuid.New()})

	for i, err := range errs {
		if !errors.Is(err, errDatabase) {
			t.Fatalf("caller %d: expected the batch error, got %v", i, err)
		}
	}
}

func TestCardsByDeck_CachesWithinOneLoader(t *testing.T) {
	deckID := uuid.New()
	lister := &mockCardLister{cards: map[uuid.UUID][]*domain.Card{deckID: {card(t, deckID, "猫")}}}
	l := New(lister, uuid.New(), dataloadgen.WithWait(time.Millisecond))

	for range 3 {
		if _, err := l.CardsByDeck.Load(context.Background(), deckID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if len(lister.batches) != 1 {
		t.Fatalf("expected repeated loads of one key to hit the lister once, got %d calls", len(lister.batches))
	}
}

func TestMiddleware_FreshLoadersPerRequestOnlyWhenAuthenticated(t *testing.T) {
	var seen []*Loaders
	handler := Middleware(&mockCardLister{}, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = append(seen, For(r.Context()))
	}))

	serve := func(authenticated bool) {
		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		if authenticated {
			req = req.WithContext(auth.WithUserID(req.Context(), uuid.New()))
		}
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}
	serve(true)
	serve(true)
	serve(false)

	if seen[0] == nil || seen[1] == nil {
		t.Fatal("expected authenticated requests to get loaders")
	}
	if seen[0] == seen[1] {
		t.Fatal("expected each request to get its own loaders; a shared cache would leak data between users")
	}
	if seen[2] != nil {
		t.Fatal("expected unauthenticated requests to get no loaders")
	}
}
