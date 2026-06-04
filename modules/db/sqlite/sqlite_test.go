package sqlite

import "testing"

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "db/sqlite" {
		t.Fatalf("expected module name %q, got %q", "db/sqlite", mod.Name())
	}
}
