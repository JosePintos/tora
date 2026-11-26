package resolvers

import (
	"context"
	"errors"
	"testing"

	deckservices "github.com/JosePintos/tora/apps/api/internal/deck/application/services"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/auth"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/model"
	"github.com/JosePintos/tora/apps/api/internal/platform/bootstrap"
	"github.com/google/uuid"
)

type MockDeckRepository struct {
	CreateFunc    func(ctx context.Context, deck *domain.Deck) error
	CreatedCalled bool
}

func (m *MockDeckRepository) Create(
	ctx context.Context,
	deck *domain.Deck,
) error {
	if m.CreateFunc != nil {
		m.CreatedCalled = true
		return m.CreateFunc(ctx, deck)
	}
	return nil
}

func (m *MockDeckRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Deck, error) {
	return nil, nil
}

func (m *MockDeckRepository) ListByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domain.Deck, error) {
	return nil, nil
}

func (m *MockDeckRepository) ListByOwner(
	ctx context.Context,
	ownerID uuid.UUID,
) ([]*domain.Deck, error) {
	return nil, nil
}

func (m *MockDeckRepository) Update(
	ctx context.Context,
	deck *domain.Deck,
) error {
	return nil
}

func (m *MockDeckRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return nil
}

func TestCreateDeck_Authenticated(t *testing.T) {
	userID := uuid.New()

	var receivedDeck *domain.Deck

	repo := &MockDeckRepository{
		CreateFunc: func(ctx context.Context, deck *domain.Deck) error {
			receivedDeck = deck
			return nil
		},
	}

	deckService := deckservices.NewDeckService(repo)

	container := &bootstrap.Container{
		DeckService: deckService,
	}

	resolver := &mutationResolver{Resolver: &Resolver{Container: container}}

	ctx := auth.WithUserID(context.Background(), userID)

	input := model.CreateDeckInput{
		Name: "JLPT N5",
	}

	result, err := resolver.CreateDeck(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected deck, got nil")
	}

	if receivedDeck == nil {
		t.Fatal("expected repository Create to be called")
	}

	if receivedDeck.OwnerID != userID {
		t.Fatalf(
			"expected owner ID %v, got %v",
			userID,
			receivedDeck.OwnerID,
		)
	}

	if receivedDeck.Name != "JLPT N5" {
		t.Fatalf(
			"expected name %q, got %q",
			"JLPT N5",
			receivedDeck.Name,
		)
	}
}

func TestCreateDeck_Unauthenticated(t *testing.T) {
	repo := &MockDeckRepository{}

	deckService := deckservices.NewDeckService(repo)

	container := &bootstrap.Container{
		DeckService: deckService,
	}

	resolver := &mutationResolver{Resolver: &Resolver{Container: container}}

	ctx := context.Background() // No user ID in context

	input := model.CreateDeckInput{
		Name: "JLPT N5",
	}

	result, err := resolver.CreateDeck(ctx, input)

	if err == nil {
		t.Fatal("expected unauthenticated error, got nil")
	}

	if !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("expected ErrUnauthenticated, got %v", err)
	}

	if result != nil {
		t.Fatal("expected nil deck, got a result")
	}

	if repo.CreatedCalled {
		t.Fatal("repository Create should not have been called")
	}
}
