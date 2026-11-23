// Package errorpresenter turns application errors into GraphQL errors.
//
// It is the single place where the API decides what a client sees when something
// fails. Classified errors (see package apperror) keep their message and get a
// stable extensions.code. Anything unclassified is logged and replaced with a
// generic message, so database or driver details never reach clients.
package errorpresenter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JosePintos/tora/apps/api/internal/apperror"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Codes sent in extensions.code. They are part of the API contract: clients
// branch on them, so don't rename them. The names follow Apollo's conventions,
// which most GraphQL client tooling understands.
const (
	CodeBadUserInput        = "BAD_USER_INPUT"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeUnauthenticated     = "UNAUTHENTICATED"
	CodeForbidden           = "FORBIDDEN"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
)

const internalMessage = "internal server error"

var codeByKind = map[apperror.Kind]string{
	apperror.Invalid:         CodeBadUserInput,
	apperror.NotFound:        CodeNotFound,
	apperror.Conflict:        CodeConflict,
	apperror.Unauthenticated: CodeUnauthenticated,
	apperror.Forbidden:       CodeForbidden,
	apperror.Internal:        CodeInternalServerError,
}

// New returns a gqlgen error presenter that logs unclassified errors with logger.
func New(logger *slog.Logger) graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		// gqlgen hands us resolver errors already wrapped in a *gqlerror.Error
		// (with the field path set); the resolver's own error is in .Err.
		presented := graphql.DefaultErrorPresenter(ctx, err)

		// Errors gqlgen produces itself (parse/validation errors, recovered
		// panics) either carry a code already or have no underlying error.
		// Their messages are written for clients, so pass them through.
		if _, hasCode := presented.Extensions["code"]; hasCode || presented.Err == nil {
			return presented
		}

		var appErr *apperror.Error
		if errors.As(presented.Err, &appErr) && appErr.Kind() != apperror.Internal {
			// Use the classified error's own message, not err.Error(): wrapping
			// context such as "create deck: ..." is for logs, not clients.
			presented.Message = appErr.Error()
			setCode(presented, codeByKind[appErr.Kind()])
			return presented
		}

		// Unclassified: this is our bug or an infrastructure failure. Log the
		// real error once, here, and hide it from the client.
		logger.ErrorContext(ctx, "graphql request failed",
			"error", presented.Err,
			"path", presented.Path.String(),
		)
		presented.Message = internalMessage
		setCode(presented, CodeInternalServerError)
		return presented
	}
}

// Recover returns a gqlgen recover function that logs panics with their stack
// trace instead of printing them to stderr (gqlgen's default).
func Recover(logger *slog.Logger) graphql.RecoverFunc {
	return func(ctx context.Context, p any) error {
		logger.ErrorContext(ctx, "panic in graphql resolver",
			"panic", fmt.Sprint(p),
			"stack", string(debug.Stack()),
		)

		gqlErr := gqlerror.Errorf(internalMessage)
		setCode(gqlErr, CodeInternalServerError)
		return gqlErr
	}
}

func setCode(err *gqlerror.Error, code string) {
	if err.Extensions == nil {
		err.Extensions = map[string]any{}
	}
	err.Extensions["code"] = code
}
