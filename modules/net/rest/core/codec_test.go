package core

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeJSONBody(t *testing.T) {
	t.Parallel()

	value, err := decodeJSONBody([]byte(`{"id":1,"ratio":1.5,"items":[2,"two"]}`))
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if got := field(t, value, "id"); got != runtime.NewInt(1) {
		t.Fatalf("expected integer id, got %s", got.String())
	}
	if got := field(t, value, "ratio"); got != runtime.NewFloat(1.5) {
		t.Fatalf("expected float ratio, got %s", got.String())
	}

	items := field(t, value, "items")
	list, ok := items.(runtime.List)
	if !ok {
		t.Fatalf("expected items list, got %T", items)
	}
	length, err := list.Length(t.Context())
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}
	if length != 2 {
		t.Fatalf("expected 2 items, got %d", length)
	}
}

func TestDecodeJSONBodyEmpty(t *testing.T) {
	t.Parallel()

	value, err := decodeJSONBody([]byte(" \n\t "))
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if value != runtime.None {
		t.Fatalf("expected NONE, got %s", value.String())
	}
}
