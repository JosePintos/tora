package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type MockTokenProvider struct {
	ParseFn func(string) (uuid.UUID, error)
}

func (m *MockTokenProvider) GenerateToken(userID uuid.UUID) (string, error) {
	return "", nil
}

func (m *MockTokenProvider) ParseToken(token string) (uuid.UUID, error) {
	if m.ParseFn != nil {
		return m.ParseFn(token)
	}
	return uuid.Nil, nil
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	expectedUserID := uuid.New()

	tokenProvider := &MockTokenProvider{
		ParseFn: func(token string) (uuid.UUID, error) {
			if token != "valid-token" {
				t.Fatalf("unexpected token: %s", token)
			}
			return expectedUserID, nil
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected user ID in context")
		}

		if userID != expectedUserID {
			t.Fatalf("expected %v, got %v", expectedUserID, userID)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenProvider, next)

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	tokenProvider := &MockTokenProvider{
		ParseFn: func(token string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("invalid token")
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenProvider, next)

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	tokenProvider := &MockTokenProvider{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenProvider, next)

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "MalformedHeader")

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	tokenProvider := &MockTokenProvider{}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := AuthMiddleware(tokenProvider, next)

	req := httptest.NewRequest(http.MethodPost, "/query", nil)

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
