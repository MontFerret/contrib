package core

import (
	"errors"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestResolveContent(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		content, err := ResolveContent(runtime.NewString("title = \"Ferret\""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if content.String() != "title = \"Ferret\"" {
			t.Fatalf("expected string content, got %q", content.String())
		}
	})

	t.Run("binary input", func(t *testing.T) {
		content, err := ResolveContent(runtime.NewBinary([]byte("title = \"Ferret\"")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if content.String() != "title = \"Ferret\"" {
			t.Fatalf("expected binary content to convert to string, got %q", content.String())
		}
	})

	t.Run("invalid input type", func(t *testing.T) {
		_, err := ResolveContent(runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected type error")
		}

		if !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected invalid type error, got %v", err)
		}
	})
}
