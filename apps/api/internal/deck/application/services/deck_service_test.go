package services

import (
	"context"
	"errors"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/deck/application/commands"
	"github.com/JosePintos/tora/apps/api/internal/deck/application/ports"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/google/uuid"
)

type MockDeckRepository struct {
	GetByIDFn   func(context.Context, uuid.UUID) (*domain.Deck, error)
	CreateFn    func(context.Context, *domain.Deck) error
	CreatedDeck *domain.Deck
	CreateCalls int
}

func (m *MockDeckRepository) Create(ctx context.Context, deck *domain.Deck) error {
	m.CreateCalls++
	m.CreatedDeck = deck
	if m.CreateFn != nil {
		return m.CreateFn(ctx, deck)
	}
	return nil
}

func (m *MockDeckRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deck, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockDeckRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Deck, error) {
	panic("not implemented")
}

func (m *MockDeckRepository) Update(ctx context.Context, deck *domain.Deck) error {
	panic("not implemented")
}

func (m *MockDeckRepository) Delete(ctx context.Context, id uuid.UUID) error {
	panic("not implemented")
}

func newTestDeckService(repo ports.DeckRepository) *DeckService {
	return NewDeckService(repo)
}

func TestCreateDeck_Success(t *testing.T) {
	ownerID := uuid.New()
	var savedDeck *domain.Deck

	repo := &MockDeckRepository{
		CreateFn: func(ctx context.Context, deck *domain.Deck) error {
			savedDeck = deck
			return nil
		},
	}

	service := newTestDeckService(repo)

	deck, err := service.Create(context.Background(), commands.CreateDeckCommand{
		OwnerID: ownerID,
		Name:    "Test Deck",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deck == nil {
		t.Fatalf("expected deck to be created, got nil")
	}
	if deck.Name != "Test Deck" {
		t.Errorf("expected deck name to be 'Test Deck', got %s", deck.Name)
	}
	if repo.CreateCalls != 1 {
		t.Fatalf("expected Create to be called once, got %d", repo.CreateCalls)
	}
	if repo.CreatedDeck == nil {
		t.Fatal("expected repository to receive deck")
	}
	if savedDeck == nil {
		t.Fatal("expected savedDeck to be set")
	}
	if savedDeck.OwnerID != ownerID {
		t.Fatalf("expected savedDeck owner ID to be %v, got %v", ownerID, savedDeck.OwnerID)
	}
}

func TestCreateDeck_EmptyName(t *testing.T) {
	repo := &MockDeckRepository{
		CreateFn: func(ctx context.Context, deck *domain.Deck) error {
			return nil
		},
	}
	service := newTestDeckService(repo)

	_, err := service.Create(context.Background(), commands.CreateDeckCommand{
		OwnerID: uuid.New(),
		Name:    "",
	})
	if !errors.Is(err, domain.ErrEmptyDeckName) {
		t.Fatalf("expected ErrEmptyDeckName, got %v", err)
	}
	if repo.CreateCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.CreateCalls)
	}
}

func TestCreateDeck_InvalidName(t *testing.T) {
	repo := &MockDeckRepository{
		CreateFn: func(ctx context.Context, deck *domain.Deck) error {
			return nil
		},
	}
	service := newTestDeckService(repo)
	longName := "This is a very long deck name that exceeds the maximum allowed length of one hundred characters. It should trigger an error."
	_, err := service.Create(context.Background(), commands.CreateDeckCommand{
		OwnerID: uuid.New(),
		Name:    longName,
	})
	if err == nil {
		t.Fatalf("expected error for long deck name, got nil")
	}
	if repo.CreateCalls != 0 {
		t.Fatalf("expected repository not to be called, got %d calls", repo.CreateCalls)
	}
}

func TestCreateDeck_RepositoryError(t *testing.T) {
	repo := &MockDeckRepository{
		CreateFn: func(ctx context.Context, deck *domain.Deck) error {
			return errors.New("repository error")
		},
	}

	service := newTestDeckService(repo)

	deck, err := service.Create(context.Background(), commands.CreateDeckCommand{
		OwnerID: uuid.New(),
		Name:    "Test Deck",
	})
	if err == nil {
		t.Fatalf("expected error from repository, got nil")
	}
	if deck != nil {
		t.Fatal("expected nil deck on repository error")
	}
}

func TestGetOwned(t *testing.T) {
	owner := uuid.New()
	deck, err := domain.NewDeck(owner, "JLPT N5")
	if err != nil {
		t.Fatalf("failed to build deck: %v", err)
	}
	errDatabase := errors.New("connection refused")

	tests := []struct {
		name     string
		userID   uuid.UUID
		repoDeck *domain.Deck
		repoErr  error
		wantDeck bool
		wantErr  error
	}{
		{name: "owner gets the deck", userID: owner, repoDeck: deck, wantDeck: true},
		{name: "another user's deck looks missing", userID: uuid.New(), repoDeck: deck, wantErr: ErrDeckNotFound},
		{name: "missing deck", userID: owner, repoDeck: nil, wantErr: ErrDeckNotFound},
		{name: "repository failure passes through", userID: owner, repoErr: errDatabase, wantErr: errDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockDeckRepository{
				GetByIDFn: func(context.Context, uuid.UUID) (*domain.Deck, error) {
					return tt.repoDeck, tt.repoErr
				},
			}

			got, err := newTestDeckService(repo).GetOwned(context.Background(), tt.userID, deck.ID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if (got != nil) != tt.wantDeck {
				t.Fatalf("expected deck returned = %v, got %v", tt.wantDeck, got)
			}
		})
	}
}
