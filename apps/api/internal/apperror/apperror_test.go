package apperror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JosePintos/tora/apps/api/internal/apperror"
)

var errSentinel = apperror.New(apperror.Conflict, "already exists")

func TestKindOf(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want apperror.Kind
	}{
		{"classified error", errSentinel, apperror.Conflict},
		{"wrapped classified error keeps its kind", fmt.Errorf("create user: %w", errSentinel), apperror.Conflict},
		{"plain error is internal", errors.New("connection refused"), apperror.Internal},
		{"error wrapped with %v loses its kind", fmt.Errorf("create user: %v", errSentinel), apperror.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apperror.KindOf(tt.err); got != tt.want {
				t.Fatalf("KindOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSentinelStillMatchesErrorsIs(t *testing.T) {
	wrapped := fmt.Errorf("register: %w", errSentinel)

	if !errors.Is(wrapped, errSentinel) {
		t.Fatal("expected errors.Is to match the wrapped sentinel")
	}
	if errors.Is(apperror.New(apperror.Conflict, "already exists"), errSentinel) {
		t.Fatal("two errors with the same kind and message must not be equal; sentinels compare by identity")
	}
}
