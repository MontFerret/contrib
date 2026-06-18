package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "http" {
		t.Fatalf("expected module name %q, got %q", "http", mod.Name())
	}
}

func TestModuleRunsHTTPClientFromFQL(t *testing.T) {
	t.Parallel()

	var called atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)

		if r.URL.Path != "/health" {
			t.Fatalf("expected /health path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"version": "1.0.0"})
	}))
	defer server.Close()

	engine, err := ferret.New(
		ferret.WithModules(New()),
		ferret.WithRuntimeParam("baseUrl", runtime.NewString(server.URL)),
	)
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Fatalf("unexpected engine close error: %v", err)
		}
	})

	_, err = engine.Run(context.Background(), source.NewAnonymous(`
		LET api = HTTP::CLIENT({
			baseUrl: @baseUrl,
			encoding: "json"
		})
		LET res = QUERY ONE "/health" IN api USING http
		RETURN res.version == "1.0.0"
	`))
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !called.Load() {
		t.Fatal("expected HTTP server to be called")
	}
}
