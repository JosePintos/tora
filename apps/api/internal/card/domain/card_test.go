package domain_test

import (
	"errors"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/google/uuid"
)

func TestCardDomain_NewCard_Success(t *testing.T) {
	deckID := uuid.New()
	front := "Front of the card"
	back := "Back of the card"

	card, err := domain.NewCard(deckID, front, back)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if card == nil {
		t.Fatalf("expected card to be created, got nil")
	}
	if card.DeckID != deckID {
		t.Errorf("expected DeckID %v, got %v", deckID, card.DeckID)
	}
	if card.Front != front {
		t.Errorf("expected Front %v, got %v", front, card.Front)
	}
	if card.Back != back {
		t.Errorf("expected Back %v, got %v", back, card.Back)
	}
}

func TestCardDomain_NewCard_EmptyFront(t *testing.T) {
	deckID := uuid.New()
	front := ""
	back := "Back of the card"

	card, err := domain.NewCard(deckID, front, back)
	if !errors.Is(err, domain.ErrEmptyCardFront) {
		t.Fatalf("expected error %v, got %v", domain.ErrEmptyCardFront, err)
	}
	if card != nil {
		t.Fatalf("expected card to be nil, got %v", card)
	}
}

func TestCardDomain_NewCard_EmptyBack(t *testing.T) {
	deckID := uuid.New()
	front := "Front of the card"
	back := ""

	card, err := domain.NewCard(deckID, front, back)
	if !errors.Is(err, domain.ErrEmptyCardBack) {
		t.Fatalf("expected error %v, got %v", domain.ErrEmptyCardBack, err)
	}
	if card != nil {
		t.Fatalf("expected card to be nil, got %v", card)
	}
}

func TestCardDomain_NewCard_WhiteSpaceFront(t *testing.T) {
	deckID := uuid.New()
	front := "   "
	back := "Back of the card"

	card, err := domain.NewCard(deckID, front, back)
	if !errors.Is(err, domain.ErrEmptyCardFront) {
		t.Fatalf("expected error %v, got %v", domain.ErrEmptyCardFront, err)
	}
	if card != nil {
		t.Fatalf("expected card to be nil, got %v", card)
	}
}

func TestCardDomain_NewCard_TrailingFront(t *testing.T) {
	deckID := uuid.New()
	front := "  Front of the card  "
	back := "Back of the card"

	card, err := domain.NewCard(deckID, front, back)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if card == nil {
		t.Fatalf("expected card to be created, got nil")
	}
	expectedFront := "Front of the card"
	if card.Front != expectedFront {
		t.Errorf("expected Front %v, got %v", expectedFront, card.Front)
	}
}
