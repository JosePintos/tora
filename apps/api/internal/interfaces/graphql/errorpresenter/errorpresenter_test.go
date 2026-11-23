package errorpresenter_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/errorpresenter"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var errTaken = apperror.New(apperror.Conflict, "username is already taken")

var testPath = ast.Path{ast.PathName("register")}

// present runs err through the presenter the way gqlgen does: resolver errors
// arrive wrapped in a *gqlerror.Error carrying the field path.
func present(t *testing.T, err error) (*gqlerror.Error, string) {
	t.Helper()

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	presented := errorpresenter.New(logger)(context.Background(), gqlerror.WrapPath(testPath, err))
	return presented, logs.String()
}

func TestPresenter_ClassifiedError(t *testing.T) {
	got, logs := present(t, fmt.Errorf("register: %w", errTaken))

	if got.Extensions["code"] != errorpresenter.CodeConflict {
		t.Fatalf("expected code %s, got %v", errorpresenter.CodeConflict, got.Extensions["code"])
	}
	if got.Message != "username is already taken" {
		t.Fatalf("expected the classified message without wrapping context, got %q", got.Message)
	}
	if got.Path.String() != "register" {
		t.Fatalf("expected path to be preserved, got %q", got.Path.String())
	}
	if logs != "" {
		t.Fatalf("expected client errors not to be logged, got %q", logs)
	}
}

func TestPresenter_UnclassifiedErrorIsHiddenAndLogged(t *testing.T) {
	got, logs := present(t, errors.New(`duplicate key value violates unique constraint "users_username_key"`))

	if got.Extensions["code"] != errorpresenter.CodeInternalServerError {
		t.Fatalf("expected code %s, got %v", errorpresenter.CodeInternalServerError, got.Extensions["code"])
	}
	if got.Message != "internal server error" {
		t.Fatalf("expected generic message, got %q", got.Message)
	}
	if !strings.Contains(logs, "users_username_key") {
		t.Fatalf("expected the real error to be logged, got %q", logs)
	}
}

func TestPresenter_ExplicitInternalKindIsHidden(t *testing.T) {
	got, _ := present(t, apperror.New(apperror.Internal, "secret detail"))

	if got.Message != "internal server error" {
		t.Fatalf("expected generic message, got %q", got.Message)
	}
}

func TestPresenter_PassesThroughGqlgenErrors(t *testing.T) {
	// Errors gqlgen builds itself (for example validation errors) have no
	// underlying Err and may already carry a code.
	validation := gqlerror.Errorf(`Cannot query field "nope" on type "Query".`)
	validation.Extensions = map[string]any{"code": "GRAPHQL_VALIDATION_FAILED"}

	got := errorpresenter.New(slog.Default())(context.Background(), validation)

	if got.Message != validation.Message || got.Extensions["code"] != "GRAPHQL_VALIDATION_FAILED" {
		t.Fatalf("expected gqlgen error unchanged, got %+v", got)
	}
}

func TestRecover_ReturnsGenericErrorAndLogsPanic(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	err := errorpresenter.Recover(logger)(context.Background(), "nil map write")

	var gqlErr *gqlerror.Error
	if !errors.As(err, &gqlErr) {
		t.Fatalf("expected *gqlerror.Error, got %T", err)
	}
	if gqlErr.Message != "internal server error" || gqlErr.Extensions["code"] != errorpresenter.CodeInternalServerError {
		t.Fatalf("unexpected recovered error: %+v", gqlErr)
	}
	if !strings.Contains(logs.String(), "nil map write") {
		t.Fatalf("expected panic value to be logged, got %q", logs.String())
	}
}
