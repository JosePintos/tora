package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	testdb "github.com/JosePintos/tora/apps/api/test/database"
	"github.com/JosePintos/tora/apps/api/test/fixtures"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func createDeck(t *testing.T, queries *sqlc.Queries) uuid.UUID {
	t.Helper()

	user := fixtures.CreateUser(t, queries)
	deck, err := queries.CreateDeck(context.Background(), sqlc.CreateDeckParams{
		ID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
		OwnerID:   user.ID,
		Name:      "JLPT N5",
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}
	return deck.ID.Bytes
}

func TestCreateAndListCards(t *testing.T) {
	db := testdb.New(t)
	queries := sqlc.New(db)
	repo := New(queries)

	deckID := createDeck(t, queries)
	otherDeckID := createDeck(t, queries)

	first, err := domain.NewCard(deckID, "猫", "cat")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	second, err := domain.NewCard(deckID, "犬", "dog")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	// Cards built back to back can share a timestamp at Postgres' microsecond
	// precision; make the intended order explicit so the test isn't flaky.
	second.CreatedAt = first.CreatedAt.Add(time.Second)
	other, err := domain.NewCard(otherDeckID, "鳥", "bird")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}

	for _, card := range []*domain.Card{first, second, other} {
		if err := repo.Create(context.Background(), card); err != nil {
			t.Fatalf("failed to create card: %v", err)
		}
	}

	cards, err := repo.ListByDeck(context.Background(), deckID)
	if err != nil {
		t.Fatalf("failed to list cards: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("expected 2 cards for the deck, got %d", len(cards))
	}
	if cards[0].ID != first.ID || cards[1].ID != second.ID {
		t.Fatalf("expected cards in creation order, got %v then %v", cards[0].Front, cards[1].Front)
	}
	if cards[0].Front != "猫" || cards[0].Back != "cat" || cards[0].DeckID != deckID {
		t.Fatalf("card fields not mapped correctly: %+v", cards[0])
	}
}

func TestListByDeck_EmptyDeckReturnsEmptySlice(t *testing.T) {
	db := testdb.New(t)
	queries := sqlc.New(db)

	cards, err := New(queries).ListByDeck(context.Background(), createDeck(t, queries))
	if err != nil {
		t.Fatalf("failed to list cards: %v", err)
	}
	if cards == nil || len(cards) != 0 {
		t.Fatalf("expected an empty, non-nil slice, got %#v", cards)
	}
}

func TestListByDecks(t *testing.T) {
	db := testdb.New(t)
	queries := sqlc.New(db)
	repo := New(queries)

	deckA, deckB, otherDeck := createDeck(t, queries), createDeck(t, queries), createDeck(t, queries)

	a1, err := domain.NewCard(deckA, "a1", "back")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	a2, err := domain.NewCard(deckA, "a2", "back")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	a2.CreatedAt = a1.CreatedAt.Add(time.Second)
	b1, err := domain.NewCard(deckB, "b1", "back")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	other, err := domain.NewCard(otherDeck, "other", "back")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}
	for _, card := range []*domain.Card{a2, b1, a1, other} {
		if err := repo.Create(context.Background(), card); err != nil {
			t.Fatalf("failed to create card: %v", err)
		}
	}

	cards, err := repo.ListByDecks(context.Background(), []uuid.UUID{deckA, deckB})
	if err != nil {
		t.Fatalf("failed to list cards: %v", err)
	}

	byDeck := map[uuid.UUID][]string{}
	for _, card := range cards {
		byDeck[card.DeckID] = append(byDeck[card.DeckID], card.Front)
	}
	if len(cards) != 3 || len(byDeck[otherDeck]) != 0 {
		t.Fatalf("expected only the 3 cards of the requested decks, got %v", byDeck)
	}
	if got := byDeck[deckA]; len(got) != 2 || got[0] != "a1" || got[1] != "a2" {
		t.Fatalf("expected deck A's cards oldest first, got %v", got)
	}
}
