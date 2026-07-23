package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MontFerret/ferret/v2"
	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk/sdktest"
)

func TestNewSmoke(t *testing.T) {
	mod := New()

	if mod == nil {
		t.Fatal("expected module to be non-nil")
	}

	if mod.Name() != "net/rest" {
		t.Fatalf("expected module name %q, got %q", "net/rest", mod.Name())
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

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithRuntimeParam("baseUrl", runtime.NewString(server.URL)),
		ferret.WithNetwork(newTestNetwork(t, ferrethttp.WithAllowLocalhost(true))),
	)

	output, err := harness.Run(context.Background(), `
		LET api = NET::REST::CLIENT({
			baseUrl: @baseUrl,
			encoding: "json"
		})
		LET res = QUERY ONE "/health" IN api USING http
		RETURN res.version == "1.0.0"
	`)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !called.Load() {
		t.Fatal("expected HTTP server to be called")
	}

	assertOutputBool(t, output.Content, true)
}

func TestModuleRunsQueryModifiersFromFQL(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)

		if r.URL.Path != "/users" {
			t.Fatalf("expected /users path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "Ada"},
			{"id": 2, "name": "Grace"},
		})
	}))
	defer server.Close()

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithRuntimeParam("baseUrl", runtime.NewString(server.URL)),
		ferret.WithNetwork(newTestNetwork(t, ferrethttp.WithAllowLocalhost(true))),
	)

	output, err := harness.Run(context.Background(), `
		LET api = NET::REST::CLIENT({
			baseUrl: @baseUrl,
			encoding: "json"
		})
		LET users = QUERY "/users" IN api
		LET first = QUERY ONE "/users" IN api USING http
		LET count = QUERY COUNT "/users" IN api
		LET exists = QUERY EXISTS "/users" IN api
		RETURN users[0].name == "Ada" AND first.id == 1 AND count == 2 AND exists
	`)
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if calls.Load() != 4 {
		t.Fatalf("expected 4 HTTP calls, got %d", calls.Load())
	}

	assertOutputBool(t, output.Content, true)
}

func TestModuleHonorsFerretHTTPPolicy(t *testing.T) {
	t.Parallel()

	var called atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	fnet, err := ferretnet.New(ferretnet.WithHTTPTransport(server.Client().Transport, ferrethttp.WithAllowedHosts("allowed.example")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	harness := sdktest.New(t,
		ferret.WithModules(New()),
		ferret.WithRuntimeParam("baseUrl", runtime.NewString(server.URL)),
		ferret.WithNetwork(fnet),
	)

	_, err = harness.Run(context.Background(), `
		LET api = NET::REST::CLIENT({
			baseUrl: @baseUrl,
			encoding: "json"
		})
		RETURN QUERY ONE "/health" IN api
	`)
	if err == nil {
		t.Fatal("expected HTTP policy error")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() {
		t.Fatal("expected HTTP policy to block before outbound request")
	}
}

func assertOutputBool(t *testing.T, data []byte, expected bool) {
	t.Helper()

	var actual bool
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("failed to decode output bool: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected output %v, got %v", expected, actual)
	}
}

func newTestNetwork(t testing.TB, policies ...ferrethttp.PolicyOption) ferretnet.Network {
	t.Helper()

	network, err := ferretnet.New(ferretnet.WithHTTPPolicies(policies...))
	if err != nil {
		t.Fatalf("failed to create test network: %v", err)
	}

	return network
}
