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

	cardservices "github.com/JosePintos/tora/apps/api/internal/card/application/services"
	carddomain "github.com/JosePintos/tora/apps/api/internal/card/domain"
	"github.com/JosePintos/tora/apps/api/internal/card/infrastructure/deckaccess"
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
	CreateFn  func(ctx context.Context, deck *deckdomain.Deck) error
	GetByIDFn func(ctx context.Context, id uuid.UUID) (*deckdomain.Deck, error)
}

func (m *mockDeckRepository) Create(ctx context.Context, deck *deckdomain.Deck) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, deck)
	}
	return nil
}

func (m *mockDeckRepository) GetByID(ctx context.Context, id uuid.UUID) (*deckdomain.Deck, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDeckRepository) ListByOwner(context.Context, uuid.UUID) ([]*deckdomain.Deck, error) {
	return nil, nil
}

func (m *mockDeckRepository) Update(context.Context, *deckdomain.Deck) error { return nil }

func (m *mockDeckRepository) Delete(context.Context, uuid.UUID) error { return nil }

type mockUserRepository struct {
	CreateFn        func(ctx context.Context, user *userdomain.User) error
	GetByUsernameFn func(ctx context.Context, username string) (*userdomain.User, error)
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

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*userdomain.User, error) {
	if m.GetByUsernameFn != nil {
		return m.GetByUsernameFn(ctx, username)
	}
	return nil, nil
}

// mockPasswordHasher accepts only the password "correct".
type mockPasswordHasher struct{}

func (mockPasswordHasher) Hash(password string) (string, error) { return "hash", nil }

func (mockPasswordHasher) Compare(hash, password string) error {
	if password != "correct" {
		return errors.New("mismatch")
	}
	return nil
}

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
			name:        "login with unknown username",
			query:       `mutation { login(input: { username: "nobody", password: "wrong" }) { token } }`,
			wantCode:    "UNAUTHENTICATED",
			wantMessage: "invalid username or password",
		},
		{
			name:  "login with wrong password looks identical",
			query: `mutation { login(input: { username: "alice", password: "wrong" }) { token } }`,
			userRepo: &mockUserRepository{GetByUsernameFn: func(context.Context, string) (*userdomain.User, error) {
				return &userdomain.User{ID: uuid.New(), Username: "alice", PasswordHash: "hash"}, nil
			}},
			wantCode:    "UNAUTHENTICATED",
			wantMessage: "invalid username or password",
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

type mockCardRepository struct {
	created []*carddomain.Card
}

func (m *mockCardRepository) Create(_ context.Context, card *carddomain.Card) error {
	m.created = append(m.created, card)
	return nil
}

func (m *mockCardRepository) ListByDeck(_ context.Context, deckID uuid.UUID) ([]*carddomain.Card, error) {
	var result []*carddomain.Card
	for _, card := range m.created {
		if card.DeckID == deckID {
			result = append(result, card)
		}
	}
	return result, nil
}

// newCardTestRouter wires the real deck and card services (and the real
// deckaccess adapter between them) over in-memory repositories. The token
// "valid" authenticates as caller; decks maps each existing deck to its owner.
func newCardTestRouter(caller uuid.UUID, decks map[uuid.UUID]uuid.UUID, cards *mockCardRepository) http.Handler {
	deckRepo := &mockDeckRepository{GetByIDFn: func(_ context.Context, id uuid.UUID) (*deckdomain.Deck, error) {
		owner, ok := decks[id]
		if !ok {
			return nil, nil
		}
		return &deckdomain.Deck{ID: id, OwnerID: owner, Name: "deck"}, nil
	}}
	tokens := mockTokenProvider{userID: caller}

	deckService := deckservices.NewDeckService(deckRepo)
	container := &bootstrap.Container{
		UserService:   userservices.NewUserService(&mockUserRepository{}, mockPasswordHasher{}, tokens),
		DeckService:   deckService,
		CardService:   cardservices.NewCardService(cards, deckaccess.New(deckService)),
		TokenProvider: tokens,
	}

	return NewRouter(resolvers.New(container), slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
}

func TestRouter_CardErrors(t *testing.T) {
	caller, stranger := uuid.New(), uuid.New()
	ownDeck, strangersDeck, missingDeck := uuid.New(), uuid.New(), uuid.New()
	decks := map[uuid.UUID]uuid.UUID{ownDeck: caller, strangersDeck: stranger}

	createCard := func(deckID string) string {
		return `mutation { createCard(input: { deckId: "` + deckID + `", front: "猫", back: "cat" }) { id } }`
	}

	tests := []struct {
		name        string
		token       string
		query       string
		wantCode    string
		wantMessage string
	}{
		{"create without token", "", createCard(ownDeck.String()), "UNAUTHENTICATED", "unauthenticated"},
		{"create in another user's deck", "valid", createCard(strangersDeck.String()), "NOT_FOUND", "deck not found"},
		{"create in a missing deck looks the same", "valid", createCard(missingDeck.String()), "NOT_FOUND", "deck not found"},
		{"create with malformed deck id", "valid", createCard("not-a-uuid"), "BAD_USER_INPUT", "invalid id"},
		{"create with empty front", "valid", `mutation { createCard(input: { deckId: "` + ownDeck.String() + `", front: " ", back: "cat" }) { id } }`, "BAD_USER_INPUT", "card front cannot be empty"},
		{"list another user's deck", "valid", `{ cards(deckId: "` + strangersDeck.String() + `") { id } }`, "NOT_FOUND", "deck not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cards := &mockCardRepository{}
			resp := doQuery(t, newCardTestRouter(caller, decks, cards), tt.token, tt.query)
			got := resp.Errors[0]

			if got.Extensions["code"] != tt.wantCode {
				t.Errorf("code = %v, want %s", got.Extensions["code"], tt.wantCode)
			}
			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			if len(cards.created) != 0 {
				t.Errorf("expected no card to be created, got %d", len(cards.created))
			}
		})
	}
}

func TestRouter_CreateAndListCards(t *testing.T) {
	caller, deckID := uuid.New(), uuid.New()
	cards := &mockCardRepository{}
	router := newCardTestRouter(caller, map[uuid.UUID]uuid.UUID{deckID: caller}, cards)

	post := func(query string) map[string]any {
		t.Helper()
		body, _ := json.Marshal(map[string]string{"query": query})
		req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer valid")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var resp map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response %q: %v", rec.Body.String(), err)
		}
		if resp["errors"] != nil {
			t.Fatalf("unexpected errors: %s", rec.Body.String())
		}
		return resp["data"].(map[string]any)
	}

	created := post(`mutation { createCard(input: { deckId: "` + deckID.String() + `", front: " 猫 ", back: "cat" }) { id deckId front back } }`)["createCard"].(map[string]any)
	if created["deckId"] != deckID.String() || created["front"] != "猫" || created["back"] != "cat" {
		t.Fatalf("unexpected created card: %v", created)
	}

	listed := post(`{ cards(deckId: "` + deckID.String() + `") { id front } }`)["cards"].([]any)
	if len(listed) != 1 || listed[0].(map[string]any)["id"] != created["id"] {
		t.Fatalf("expected the created card to be listed, got %v", listed)
	}
}
