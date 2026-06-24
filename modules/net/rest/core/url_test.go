package core

import (
	"context"
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestResolveRequestURLMergesQueryValues(t *testing.T) {
	t.Parallel()

	got, err := resolveRequestURL(context.Background(), "https://api.example.test/v1/", "/users?existing=1", object(t, map[string]runtime.Value{
		"active": runtime.True,
		"limit":  runtime.NewInt(50),
		"none":   runtime.None,
		"tag":    runtime.NewArrayWith(runtime.NewString("admin"), runtime.NewString("active")),
	}))
	if err != nil {
		t.Fatalf("unexpected URL error: %v", err)
	}

	want := "https://api.example.test/users?active=true&existing=1&limit=50&tag=admin&tag=active"
	if got != want {
		t.Fatalf("unexpected URL:\nwant %s\n got %s", want, got)
	}
}

func TestResolveRequestURLRejectsNestedQueryObjects(t *testing.T) {
	t.Parallel()

	_, err := resolveRequestURL(context.Background(), "https://api.example.test", "/users", object(t, map[string]runtime.Value{
		"filter": object(t, map[string]runtime.Value{"active": runtime.True}),
	}))
	if err == nil {
		t.Fatal("expected nested query object error")
	}
	if !strings.Contains(err.Error(), "nested objects are not supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}
