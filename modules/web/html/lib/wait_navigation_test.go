package lib

import (
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParseWaitNavigationParams(t *testing.T) {
	params, err := parseWaitNavigationParams(t.Context(), runtime.NewObjectWith(map[string]runtime.Value{
		"target":  runtime.NewString("https://example.com"),
		"timeout": runtime.NewInt(1234),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.TargetURL != "https://example.com" {
		t.Fatalf("expected target URL to decode, got %q", params.TargetURL)
	}

	if params.Timeout != 1234 {
		t.Fatalf("expected timeout to decode, got %d", params.Timeout)
	}

	_, err = parseWaitNavigationParams(t.Context(), runtime.NewObjectWith(map[string]runtime.Value{
		"target": runtime.NewString("https://example.com"),
		"extra":  runtime.True,
	}))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
