package toml

import "testing"

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "toml" {
		t.Fatalf("expected module name %q, got %q", "toml", mod.Name())
	}
}
