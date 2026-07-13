package domain

import (
	"testing"

	"github.com/google/uuid"
)

// You want to verify the business invariants.
func TestNewUser(t *testing.T) {
	user := NewUser("testuser", "password123")
	if user.Username != "testuser" {
		t.Fatal("expected username to be testuser")
	}

	if user.PasswordHash != "password123" {
		t.Fatal("expected password hash to be set")
	}

	if user.ID == uuid.Nil {
		t.Fatal("expected ID to be generated")
	}

	if user.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}
