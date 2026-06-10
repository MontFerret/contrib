package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestBuildResults(t *testing.T) {
	ctx := context.Background()
	parsed := &parsedToken{
		rawHeader:    "h",
		rawPayload:   "p",
		rawSignature: "s",
		header:       map[string]any{"alg": "HS256"},
		claims:       map[string]any{"sub": "user123"},
	}

	t.Run("Inspect result", func(t *testing.T) {
		res, err := buildInspectResult(parsed)
		if err != nil {
			t.Fatalf("buildInspectResult() error = %v", err)
		}
		obj := res.(*runtime.Object)
		val, _ := obj.Get(ctx, runtime.NewString("verified"))
		if val != runtime.False {
			t.Errorf("verified should be false")
		}
	})

	t.Run("Verify result", func(t *testing.T) {
		res, err := buildVerifyResult(parsed)
		if err != nil {
			t.Fatalf("buildVerifyResult() error = %v", err)
		}
		obj := res.(*runtime.Object)
		val, _ := obj.Get(ctx, runtime.NewString("verified"))
		if val != runtime.True {
			t.Errorf("verified should be true")
		}
	})
}
