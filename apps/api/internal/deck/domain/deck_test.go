package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewDeck(t *testing.T) {
	ownerID := uuid.New()
	deck, err := NewDeck(ownerID, "Test Deck")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deck.Name != "Test Deck" {
		t.Fatalf("expected deck name to be 'Test Deck', got %s", deck.Name)
	}
	if deck.OwnerID != ownerID {
		t.Fatalf("expected owner ID to be %v, got %v", ownerID, deck.OwnerID)
	}
	if deck.ID == uuid.Nil {
		t.Fatal("expected ID to be generated")
	}
}

func TestNewDeck_EmptyName(t *testing.T) {
	ownerID := uuid.New()
	_, err := NewDeck(ownerID, "")
	if !errors.Is(err, ErrEmptyDeckName) {
		t.Fatalf("expected ErrEmptyDeckName, got %v", err)
	}
}

func TestNewDeck_LongName(t *testing.T) {
	ownerID := uuid.New()
	longName := "This is a very long deck name that exceeds the maximum allowed length of one hundred characters. It should trigger an error."
	_, err := NewDeck(ownerID, longName)
	if !errors.Is(err, ErrDeckNameTooLong) {
		t.Fatalf("expected ErrDeckNameTooLong, got %v", err)
	}
}

func TestRenameDeck(t *testing.T) {
	ownerID := uuid.New()
	deck, err := NewDeck(ownerID, "Initial Name")
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}
	renameErr := deck.Rename("New Name")
	if renameErr != nil {
		t.Fatalf("expected no error when renaming, got %v", renameErr)
	}
}

func TestRenameDeck_EmptyName(t *testing.T) {
	ownerID := uuid.New()
	deck, err := NewDeck(ownerID, "Initial Name")
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}
	emptyNameErr := deck.Rename("")
	if !errors.Is(emptyNameErr, ErrEmptyDeckName) {
		t.Fatalf("expected ErrEmptyDeckName, got %v", emptyNameErr)
	}
}

func TestRenameDeck_LongName(t *testing.T) {
	ownerID := uuid.New()
	deck, err := NewDeck(ownerID, "Initial Name")
	if err != nil {
		t.Fatalf("failed to create deck: %v", err)
	}
	longName := "This is a very long deck name that exceeds the maximum allowed length of one hundred characters. It should trigger an error."
	longNameErr := deck.Rename(longName)
	if !errors.Is(longNameErr, ErrDeckNameTooLong) {
		t.Fatalf("expected ErrDeckNameTooLong, got %v", longNameErr)
	}
}

func TestNewDeck_NameLengthCountsCharactersNotBytes(t *testing.T) {
	// 100 kanji = 300 bytes in UTF-8, but only 100 characters.
	name := strings.Repeat("漢", 100)

	if _, err := NewDeck(uuid.New(), name); err != nil {
		t.Fatalf("expected 100-character Japanese name to be valid, got %v", err)
	}

	if _, err := NewDeck(uuid.New(), name+"字"); !errors.Is(err, ErrDeckNameTooLong) {
		t.Fatalf("expected ErrDeckNameTooLong for 101 characters, got %v", err)
	}
}

func TestDeck_IsOwnedBy(t *testing.T) {
	owner := uuid.New()
	deck, err := NewDeck(owner, "JLPT N5")
	if err != nil {
		t.Fatalf("failed to build deck: %v", err)
	}

	if !deck.IsOwnedBy(owner) {
		t.Fatal("expected the creator to own the deck")
	}
	if deck.IsOwnedBy(uuid.New()) {
		t.Fatal("expected another user not to own the deck")
	}
}
