package dom

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDecodeDOMEventPayloadPreservesContextErrors(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	expired, stop := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer stop()

	for _, ctx := range []context.Context{canceled, expired} {
		_, err := decodeDOMEventPayload(ctx, `{"type":"click"}`)
		if !errors.Is(err, ctx.Err()) {
			t.Fatalf("expected %v, got %v", ctx.Err(), err)
		}
	}
}
