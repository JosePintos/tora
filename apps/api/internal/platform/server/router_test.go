package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	deckservices "github.com/JosePintos/tora/apps/api/internal/deck/application/services"
	deckdomain "github.com/JosePintos/tora/apps/api/internal/deck/domain"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/resolvers"
	"github.com/JosePintos/tora/apps/api/internal/platform/bootstrap"
	userservices "github.com/JosePintos/tora/apps/api/internal/user/application/services"
	userdomain "github.com/JosePintos/tora/apps/api/internal/user/domain"
	"github.com/google/uuid"
)

// These tests send real GraphQL requests through NewRouter, so they cover the
// whole error path: domain/service error -> gqlgen -> error presenter -> JSON.

type mockDeckRepository struct {
	CreateFn func(ctx context.Context, deck *deckdomain.Deck) error
}

func (m *mockDeckRepository) Create(ctx context.Context, deck *deckdomain.Deck) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, deck)
	}
	return nil
}

func (m *mockDeckRepository) GetByID(context.Context, uuid.UUID) (*deckdomain.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) ListByOwner(context.Context, uuid.UUID) ([]*deckdomain.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) Update(context.Context, *deckdomain.Deck) error { return nil }

func (m *mockDeckRepository) Delete(context.Context, uuid.UUID) error { return nil }

type mockUserRepository struct {
	CreateFn func(ctx context.Context, user *userdomain.User) error
}

func (m *mockUserRepository) Create(ctx context.Context, user *userdomain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepository) GetByID(context.Context, uuid.UUID) (*userdomain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByUsername(context.Context, string) (*userdomain.User, error) {
	return nil, nil
}

type mockPasswordHasher struct{}

func (mockPasswordHasher) Hash(password string) (string, error) { return "hash", nil }

func (mockPasswordHasher) Compare(hash, password string) error { return nil }

// mockTokenProvider accepts the token "valid" for userID.
type mockTokenProvider struct{ userID uuid.UUID }

func (m mockTokenProvider) GenerateToken(uuid.UUID) (string, error) { return "valid", nil }

func (m mockTokenProvider) ParseToken(token string) (uuid.UUID, error) {
	if token != "valid" {
		return uuid.Nil, errors.New("invalid token")
	}
	return m.userID, nil
}

type graphqlResponse struct {
	Errors []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}

func newTestRouter(deckRepo *mockDeckRepository, userRepo *mockUserRepository, logs *bytes.Buffer) http.Handler {
	tokens := mockTokenProvider{userID: uuid.New()}

	container := &bootstrap.Container{
		UserService:   userservices.NewUserService(userRepo, mockPasswordHasher{}, tokens),
		DeckService:   deckservices.NewDeckService(deckRepo),
		TokenProvider: tokens,
	}

	logger := slog.New(slog.NewTextHandler(logs, nil))
	return NewRouter(resolvers.New(container), logger)
}

func doQuery(t *testing.T, router http.Handler, token, query string) graphqlResponse {
	t.Helper()

	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		t.Fatalf("failed to encode request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp graphqlResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response %q: %v", rec.Body.String(), err)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("expected exactly one error, got %s", rec.Body.String())
	}
	return resp
}

func TestRouter_ErrorResponses(t *testing.T) {
	const createDeck = `mutation { createDeck(input: { name: "%s" }) { id } }`

	tests := []struct {
		name        string
		token       string
		query       string
		deckRepo    *mockDeckRepository
		userRepo    *mockUserRepository
		wantCode    string
		wantMessage string
		wantLogged  string // substring expected in the server log; "" means nothing logged
	}{
		{
			name:        "missing token",
			query:       strings.Replace(createDeck, "%s", "N5", 1),
			wantCode:    "UNAUTHENTICATED",
			wantMessage: "unauthenticated",
		},
		{
			name:        "domain validation error",
			token:       "valid",
			query:       strings.Replace(createDeck, "%s", "   ", 1),
			wantCode:    "BAD_USER_INPUT",
			wantMessage: "deck name cannot be empty",
		},
		{
			name:  "repository conflict",
			query: `mutation { register(input: { username: "taken", password: "password123" }) { id } }`,
			userRepo: &mockUserRepository{CreateFn: func(context.Context, *userdomain.User) error {
				return userdomain.ErrUsernameTaken
			}},
			wantCode:    "CONFLICT",
			wantMessage: "username is already taken",
		},
		{
			name:  "infrastructure failure is hidden",
			token: "valid",
			query: strings.Replace(createDeck, "%s", "N5", 1),
			deckRepo: &mockDeckRepository{CreateFn: func(context.Context, *deckdomain.Deck) error {
				return errors.New("dial tcp 10.0.0.5:5432: connection refused")
			}},
			wantCode:    "INTERNAL_SERVER_ERROR",
			wantMessage: "internal server error",
			wantLogged:  "connection refused",
		},
		{
			name:  "panic is recovered and hidden",
			token: "valid",
			query: strings.Replace(createDeck, "%s", "N5", 1),
			deckRepo: &mockDeckRepository{CreateFn: func(context.Context, *deckdomain.Deck) error {
				panic("nil map write")
			}},
			wantCode:    "INTERNAL_SERVER_ERROR",
			wantMessage: "internal server error",
			wantLogged:  "nil map write",
		},
		{
			name:        "gqlgen validation errors keep their own code",
			query:       `{ doesNotExist }`,
			wantCode:    "GRAPHQL_VALIDATION_FAILED",
			wantMessage: `Cannot query field "doesNotExist" on type "Query".`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.deckRepo == nil {
				tt.deckRepo = &mockDeckRepository{}
			}
			if tt.userRepo == nil {
				tt.userRepo = &mockUserRepository{}
			}

			var logs bytes.Buffer
			resp := doQuery(t, newTestRouter(tt.deckRepo, tt.userRepo, &logs), tt.token, tt.query)
			got := resp.Errors[0]

			if got.Extensions["code"] != tt.wantCode {
				t.Errorf("code = %v, want %s", got.Extensions["code"], tt.wantCode)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			if tt.wantLogged == "" && logs.Len() > 0 {
				t.Errorf("expected nothing logged, got %q", logs.String())
			}
			if tt.wantLogged != "" && !strings.Contains(logs.String(), tt.wantLogged) {
				t.Errorf("expected log to contain %q, got %q", tt.wantLogged, logs.String())
			}
		})
	}
}
