// Package apperror classifies failures by kind so that each layer can decide
// what an error means without knowing who produced it.
//
// The domain and application layers decide the kind ("this is invalid input",
// "this already exists"). Inbound adapters such as GraphQL translate a kind into
// their own vocabulary (an extensions.code, an HTTP status) and hide anything
// that is not classified.
//
// This package is a deliberately tiny shared kernel: it must not import any other
// package from this module.
package apperror

import "errors"

// Kind is the category of a failure. It describes what went wrong from the
// caller's point of view, not where it happened.
type Kind uint8

const (
	// Internal is the zero value on purpose: an unclassified error is treated
	// as internal, so its details are never shown to clients.
	Internal Kind = iota
	// Invalid means the input broke a business rule or validation.
	Invalid
	// NotFound means the requested resource does not exist.
	NotFound
	// Conflict means the request clashes with existing state, such as a duplicate.
	Conflict
	// Unauthenticated means the caller is not, or could not be, identified.
	Unauthenticated
	// Forbidden means the caller is identified but not allowed to do this.
	Forbidden
)

func (k Kind) String() string {
	switch k {
	case Invalid:
		return "invalid"
	case NotFound:
		return "not_found"
	case Conflict:
		return "conflict"
	case Unauthenticated:
		return "unauthenticated"
	case Forbidden:
		return "forbidden"
	default:
		return "internal"
	}
}

// Error is an error with a Kind and a message that is safe to show to clients.
//
// Declare package-level sentinels with New so callers can still use errors.Is:
//
//	var ErrEmptyDeckName = apperror.New(apperror.Invalid, "deck name cannot be empty")
type Error struct {
	kind    Kind
	message string
}

// New returns an error of the given kind. The message is shown to clients, so it
// must not contain internal details.
func New(kind Kind, message string) *Error {
	return &Error{kind: kind, message: message}
}

func (e *Error) Error() string { return e.message }

// Kind returns the error's category.
func (e *Error) Kind() Kind { return e.kind }

// KindOf returns the kind of the first *Error in err's chain, or Internal when
// there is none. Wrapping with fmt.Errorf("...: %w", err) keeps the kind.
func KindOf(err error) Kind {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.kind
	}
	return Internal
}
