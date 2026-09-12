package postgres

import (
	"context"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/JosePintos/tora/apps/api/internal/platform/database/sqlc"
	testdb "github.com/JosePintos/tora/apps/api/test/database"
	"github.com/JosePintos/tora/apps/api/test/fixtures"
	"github.com/jackc/pgx/v5/pgtype"
)

// Connect test DB
// ↓
// Run migrations
// ↓
// Create User fixture
// ↓
// Create Deck
// ↓
// Assertions
// ↓
// TRUNCATE tables (Cleanup)
// ↓
// Close pool

func TestCreateDeck(t *testing.T) {
	db := testdb.New(t)

	queries := sqlc.New(db)

	repo := New(queries)

	user := fixtures.CreateUser(t, queries)

	deck, err := domain.NewDeck(user.ID.Bytes, "Test Deck")
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	if err := repo.Create(context.Background(), deck); err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}

	savedDeck, err := queries.GetDeckByID(context.Background(), pgtype.UUID{Bytes: deck.ID, Valid: true})
	if err != nil {
		t.Fatalf("failed to get deck by ID: %v", err)
	}

	if savedDeck.Name != deck.Name {
		t.Errorf("expected deck name %s, got %s", deck.Name, savedDeck.Name)
	}
	if savedDeck.OwnerID.Bytes != deck.OwnerID {
		t.Errorf("expected owner ID %s, got %s", deck.OwnerID, savedDeck.OwnerID.Bytes)
	}

	decks, err := repo.ListByOwner(context.Background(), deck.OwnerID)
	if err != nil {
		t.Fatalf("failed to list decks by owner: %v", err)
	}

	for _, d := range decks {
		if d.ID != deck.ID {
			t.Errorf("expected deck ID %s, got %s", deck.ID, d.ID)
		}
	}
}
