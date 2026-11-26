package deckaccess

import (
	"context"
	"errors"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/card/application/ports"
	deckservices "github.com/JosePintos/tora/apps/api/internal/deck/application/services"
	"github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/google/uuid"
)

type mockDeckOwnership struct {
	GetOwnedFn     func(ctx context.Context, ownerID, deckID uuid.UUID) (*domain.Deck, error)
	GetOwnedManyFn func(ctx context.Context, ownerID uuid.UUID, deckIDs []uuid.UUID) ([]*domain.Deck, error)
}

func (m *mockDeckOwnership) GetOwnedMany(ctx context.Context, ownerID uuid.UUID, deckIDs []uuid.UUID) ([]*domain.Deck, error) {
	return m.GetOwnedManyFn(ctx, ownerID, deckIDs)
}

func (m *mockDeckOwnership) GetOwned(ctx context.Context, ownerID, deckID uuid.UUID) (*domain.Deck, error) {
	return m.GetOwnedFn(ctx, ownerID, deckID)
}

func TestEnsureOwnsDeck(t *testing.T) {
	errDatabase := errors.New("connection refused")

	tests := []struct {
		name    string
		deckErr error
		wantErr error
	}{
		{name: "owned deck", deckErr: nil, wantErr: nil},
		{name: "deck context's not-found becomes the card context's", deckErr: deckservices.ErrDeckNotFound, wantErr: ports.ErrDeckNotFound},
		{name: "other failures pass through", deckErr: errDatabase, wantErr: errDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decks := &mockDeckOwnership{GetOwnedFn: func(context.Context, uuid.UUID, uuid.UUID) (*domain.Deck, error) {
				if tt.deckErr != nil {
					return nil, tt.deckErr
				}
				return &domain.Deck{}, nil
			}}

			err := New(decks).EnsureOwnsDeck(context.Background(), uuid.New(), uuid.New())
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestEnsureOwnsDecks(t *testing.T) {
	errDatabase := errors.New("connection refused")

	tests := []struct {
		name    string
		deckErr error
		wantErr error
	}{
		{name: "all owned", deckErr: nil, wantErr: nil},
		{name: "deck context's not-found becomes the card context's", deckErr: deckservices.ErrDeckNotFound, wantErr: ports.ErrDeckNotFound},
		{name: "other failures pass through", deckErr: errDatabase, wantErr: errDatabase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decks := &mockDeckOwnership{GetOwnedManyFn: func(context.Context, uuid.UUID, []uuid.UUID) ([]*domain.Deck, error) {
				return nil, tt.deckErr
			}}

			err := New(decks).EnsureOwnsDecks(context.Background(), uuid.New(), []uuid.UUID{uuid.New()})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
