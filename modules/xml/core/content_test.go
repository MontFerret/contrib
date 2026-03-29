package core

import (
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestResolveContent(t *testing.T) {
	t.Run("accepts string input", func(t *testing.T) {
		content, err := ResolveContent(runtime.NewString("<book/>"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if content.String() != "<book/>" {
			t.Fatalf("unexpected content: %q", content.String())
		}
	})

	t.Run("accepts binary input", func(t *testing.T) {
		content, err := ResolveContent(runtime.NewBinary([]byte("<book/>")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if content.String() != "<book/>" {
			t.Fatalf("unexpected content: %q", content.String())
		}
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		_, err := ResolveContent(runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for non-text input")
		}

		if !strings.Contains(err.Error(), "String or Binary") {
			t.Fatalf("expected error to mention String or Binary, got %v", err)
		}
	})
}
