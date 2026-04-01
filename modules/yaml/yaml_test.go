package yaml

import "testing"

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "yaml" {
		t.Fatalf("expected module name %q, got %q", "yaml", mod.Name())
	}
}
