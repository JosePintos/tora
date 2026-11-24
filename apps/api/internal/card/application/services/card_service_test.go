package services

import (
	"context"
	"errors"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/card/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/card/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/google/uuid"
)

type MockCardRepository struct {
	CreateFn     func(context.Context, *domain.Card) error
	ListByDeckFn func(context.Context, uuid.UUID) ([]*domain.Card, error)
	CreateCalls  int
	ListCalls    int
}

func (m *MockCardRepository) Create(ctx context.Context, card *domain.Card) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, card)
	}
	return nil
}

func (m *MockCardRepository) ListByDeck(ctx context.Context, deckID uuid.UUID) ([]*domain.Card, error) {
	m.ListCalls++
	if m.ListByDeckFn != nil {
		return m.ListByDeckFn(ctx, deckID)
	}
	return nil, nil
}

// MockDeckAccess grants access only to the (owner, deck) pair it was built with.
type MockDeckAccess struct {
	EnsureOwnsDeckFn func(ctx context.Context, userID, deckID uuid.UUID) error
}

func (m *MockDeckAccess) EnsureOwnsDeck(ctx context.Context, userID, deckID uuid.UUID) error {
	if m.EnsureOwnsDeckFn != nil {
		return m.EnsureOwnsDeckFn(ctx, userID, deckID)
	}
	return nil
}

func ownedBy(owner, deck uuid.UUID) *MockDeckAccess {
	return &MockDeckAccess{
		EnsureOwnsDeckFn: func(_ context.Context, userID, deckID uuid.UUID) error {
			if userID != owner || deckID != deck {
				return ports.ErrDeckNotFound
			}
			return nil
		},
	}
}

func TestCreateCard_Success(t *testing.T) {
	owner, deckID := uuid.New(), uuid.New()

	var saved *domain.Card
	repo := &MockCardRepository{CreateFn: func(_ context.Context, card *domain.Card) error {
		saved = card
		return nil
	}}

	card, err := NewCardService(repo, ownedBy(owner, deckID)).Create(context.Background(), commands.CreateCardCommand{
		UserID: owner,
		DeckID: deckID,
		Front:  "  猫 ",
		Back:   "cat",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if saved == nil || saved != card {
		t.Fatal("expected the returned card to be the one saved")
	}
	if card.DeckID != deckID {
		t.Fatalf("expected deck %v, got %v", deckID, card.DeckID)
	}
	if card.Front != "猫" {
		t.Fatalf("expected trimmed front %q, got %q", "猫", card.Front)
	}
}

func TestCreateCard_DeckNotOwned(t *testing.T) {
	owner, deckID := uuid.New(), uuid.New()
	repo := &MockCardRepository{}

	_, err := NewCardService(repo, ownedBy(owner, deckID)).Create(context.Background(), commands.CreateCardCommand{
		UserID: uuid.New(), // someone else
		DeckID: deckID,
		Front:  "猫",
		Back:   "cat",
	})
	if !errors.Is(err, ports.ErrDeckNotFound) {
		t.Fatalf("expected ErrDeckNotFound, got %v", err)
	}
	if repo.CreateCalls != 0 {
		t.Fatalf("expected no card to be saved, got %d Create calls", repo.CreateCalls)
	}
}

func TestCreateCard_InvalidInputIsRejectedBeforeOwnershipCheck(t *testing.T) {
	access := &MockDeckAccess{EnsureOwnsDeckFn: func(context.Context, uuid.UUID, uuid.UUID) error {
		t.Fatal("ownership must not be checked for invalid input")
		return nil
	}}
	repo := &MockCardRepository{}

	_, err := NewCardService(repo, access).Create(context.Background(), commands.CreateCardCommand{
		UserID: uuid.New(),
		DeckID: uuid.New(),
		Front:  "   ",
		Back:   "cat",
	})
	if !errors.Is(err, domain.ErrEmptyCardFront) {
		t.Fatalf("expected ErrEmptyCardFront, got %v", err)
	}
	if repo.CreateCalls != 0 {
		t.Fatalf("expected no card to be saved, got %d Create calls", repo.CreateCalls)
	}
}

func TestCreateCard_RepositoryError(t *testing.T) {
	owner, deckID := uuid.New(), uuid.New()
	errDatabase := errors.New("connection refused")
	repo := &MockCardRepository{CreateFn: func(context.Context, *domain.Card) error { return errDatabase }}

	_, err := NewCardService(repo, ownedBy(owner, deckID)).Create(context.Background(), commands.CreateCardCommand{
		UserID: owner, DeckID: deckID, Front: "猫", Back: "cat",
	})
	if !errors.Is(err, errDatabase) {
		t.Fatalf("expected the repository error, got %v", err)
	}
}

func TestListByDeck_Success(t *testing.T) {
	owner, deckID := uuid.New(), uuid.New()
	card, err := domain.NewCard(deckID, "猫", "cat")
	if err != nil {
		t.Fatalf("failed to build card: %v", err)
	}

	var queriedDeck uuid.UUID
	repo := &MockCardRepository{ListByDeckFn: func(_ context.Context, id uuid.UUID) ([]*domain.Card, error) {
		queriedDeck = id
		return []*domain.Card{card}, nil
	}}

	cards, err := NewCardService(repo, ownedBy(owner, deckID)).ListByDeck(context.Background(), owner, deckID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cards) != 1 || cards[0] != card {
		t.Fatalf("expected the repository's cards, got %v", cards)
	}
	if queriedDeck != deckID {
		t.Fatalf("expected cards of deck %v, got %v", deckID, queriedDeck)
	}
}

func TestListByDeck_DeckNotOwned(t *testing.T) {
	owner, deckID := uuid.New(), uuid.New()
	repo := &MockCardRepository{}

	_, err := NewCardService(repo, ownedBy(owner, deckID)).ListByDeck(context.Background(), uuid.New(), deckID)
	if !errors.Is(err, ports.ErrDeckNotFound) {
		t.Fatalf("expected ErrDeckNotFound, got %v", err)
	}
	if repo.ListCalls != 0 {
		t.Fatalf("expected the repository not to be queried, got %d calls", repo.ListCalls)
	}
}
