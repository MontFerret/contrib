package sitemap

import (
	"testing"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "web/sitemap" {
		t.Fatalf("expected module name %q, got %q", "web/sitemap", mod.Name())
	}
}
