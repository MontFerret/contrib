package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestClientQueryJSONList(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Fatalf("expected path /users, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("active") != "true" {
			t.Fatalf("expected active query parameter")
		}
		if r.URL.Query().Get("limit") != "50" {
			t.Fatalf("expected limit query parameter")
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("expected default authorization header")
		}
		if r.Header.Get("X-Request-ID") != "req-1" {
			t.Fatalf("expected request override header")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "Ada"},
			{"id": 2, "name": "Grace"},
		})
	}))
	defer server.Close()

	ctx := networkContext()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Headers.Set("Authorization", "Bearer token")
	client := NewClient(cfg)

	out, err := client.Query(ctx, runtime.Query{
		Expression: runtime.NewString("/users"),
		Params: object(t, map[string]runtime.Value{
			"query": object(t, map[string]runtime.Value{
				"active": runtime.True,
				"limit":  runtime.NewInt(50),
			}),
			"headers": object(t, map[string]runtime.Value{
				"X-Request-ID": runtime.NewString("req-1"),
			}),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}

	length, err := out.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}
	if length != 2 {
		t.Fatalf("expected 2 users, got %d", length)
	}

	first, err := out.At(ctx, runtime.ZeroInt)
	if err != nil {
		t.Fatalf("unexpected first item error: %v", err)
	}
	if got := field(t, first, "name"); got != runtime.NewString("Ada") {
		t.Fatalf("expected first user Ada, got %s", got.String())
	}
}

func TestClientInfersPostAndReturnsFullResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
			t.Fatalf("expected JSON content type, got %q", contentType)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if payload["name"] != "Ada" {
			t.Fatalf("expected JSON body name Ada, got %v", payload["name"])
		}

		w.Header().Set("X-Result", "created")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Ada"})
	}))
	defer server.Close()

	ctx := networkContext()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	client := NewClient(cfg)

	out, err := client.QueryOne(ctx, runtime.Query{
		Expression: runtime.NewString("/users"),
		Params: object(t, map[string]runtime.Value{
			"body": object(t, map[string]runtime.Value{
				"name":  runtime.NewString("Ada"),
				"email": runtime.NewString("ada@example.com"),
			}),
		}),
		Options: object(t, map[string]runtime.Value{
			"response": runtime.NewString("full"),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}

	if got := field(t, out, "ok"); got != runtime.True {
		t.Fatalf("expected ok=true, got %s", got.String())
	}
	if got := field(t, out, "status"); got != runtime.NewInt(http.StatusCreated) {
		t.Fatalf("expected status 201, got %s", got.String())
	}
	if got := field(t, field(t, out, "body"), "id"); got != runtime.NewInt(1) {
		t.Fatalf("expected body id 1, got %s", got.String())
	}
	if got := field(t, field(t, out, "headers"), "x-result"); got != runtime.NewString("created") {
		t.Fatalf("expected response header, got %s", got.String())
	}
}

func TestClientErrorModes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "User not found"})
	}))
	defer server.Close()

	ctx := networkContext()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL

	_, err := NewClient(cfg).QueryOne(ctx, runtime.Query{Expression: runtime.NewString("/users/404")})
	if err == nil {
		t.Fatal("expected default raise mode to return an error")
	}
	if !strings.Contains(err.Error(), "unexpected status 404 Not Found") {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg.ErrorMode = ErrorModeResponse
	out, err := NewClient(cfg).QueryOne(ctx, runtime.Query{Expression: runtime.NewString("/users/404")})
	if err != nil {
		t.Fatalf("unexpected response mode error: %v", err)
	}
	if got := field(t, out, "ok"); got != runtime.False {
		t.Fatalf("expected ok=false, got %s", got.String())
	}
	if got := field(t, out, "status"); got != runtime.NewInt(http.StatusNotFound) {
		t.Fatalf("expected status 404, got %s", got.String())
	}
	if got := field(t, field(t, out, "body"), "error"); got != runtime.NewString("User not found") {
		t.Fatalf("expected decoded error body, got %s", got.String())
	}
}

func TestClientResponseEncodings(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/text":
			_, _ = w.Write([]byte("hello"))
		case "/bytes":
			_, _ = w.Write([]byte{0x01, 0x02, 0x03})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	ctx := networkContext()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.ResponseEncoding = EncodingText

	text, err := NewClient(cfg).QueryOne(ctx, runtime.Query{Expression: runtime.NewString("/text")})
	if err != nil {
		t.Fatalf("unexpected text query error: %v", err)
	}
	if text != runtime.NewString("hello") {
		t.Fatalf("expected text response, got %s", text.String())
	}

	bytesValue, err := NewClient(cfg).QueryOne(ctx, runtime.Query{
		Expression: runtime.NewString("/bytes"),
		Options: object(t, map[string]runtime.Value{
			"responseEncoding": runtime.NewString("bytes"),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected bytes query error: %v", err)
	}

	binary, ok := bytesValue.(runtime.Binary)
	if !ok {
		t.Fatalf("expected binary response, got %T", bytesValue)
	}
	if string(binary) != "\x01\x02\x03" {
		t.Fatalf("unexpected binary response: %v", []byte(binary))
	}
}

func TestClientFormRequestEncoding(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if contentType := r.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
			t.Fatalf("expected form content type, got %q", contentType)
		}

		body, err := url.ParseQuery(readBody(t, r))
		if err != nil {
			t.Fatalf("failed to parse form body: %v", err)
		}
		if body.Get("name") != "Ada" {
			t.Fatalf("expected name Ada, got %q", body.Get("name"))
		}
		if got := body["tag"]; len(got) != 2 || got[0] != "admin" || got[1] != "active" {
			t.Fatalf("unexpected repeated tags: %v", got)
		}

		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	ctx := networkContext()
	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	client := NewClient(cfg)

	out, err := client.QueryOne(ctx, runtime.Query{
		Expression: runtime.NewString("/users"),
		Params: object(t, map[string]runtime.Value{
			"method": runtime.NewString("POST"),
			"body": object(t, map[string]runtime.Value{
				"name": runtime.NewString("Ada"),
				"tag":  runtime.NewArrayWith(runtime.NewString("admin"), runtime.NewString("active")),
			}),
		}),
		Options: object(t, map[string]runtime.Value{
			"requestEncoding":  runtime.NewString("form"),
			"responseEncoding": runtime.NewString("text"),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected form query error: %v", err)
	}
	if out != runtime.NewString("ok") {
		t.Fatalf("expected ok response, got %s", out.String())
	}
}

func TestClientAcceptsHTTPDialect(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Fatalf("expected path /users, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1}})
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	client := NewClient(cfg)

	ctx := networkContext()
	out, err := client.Query(ctx, runtime.Query{
		Kind:       runtime.NewString("http"),
		Expression: runtime.NewString("/users"),
	})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}

	length, err := out.Length(ctx)
	if err != nil {
		t.Fatalf("unexpected length error: %v", err)
	}
	if length != 1 {
		t.Fatalf("expected 1 item, got %d", length)
	}
}

func TestClientRequestTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(time.Second):
			t.Fatal("request context was not canceled")
		}
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	cfg.Timeout = int64(10 * time.Millisecond)
	client := NewClient(cfg)

	_, err := client.QueryOne(networkContext(), runtime.Query{Expression: runtime.NewString("/slow")})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientRejectsUnsupportedDialect(t *testing.T) {
	t.Parallel()

	var called atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = server.URL
	client := NewClient(cfg)

	_, err := client.Query(context.Background(), runtime.Query{
		Kind:       runtime.NewString("sql"),
		Expression: runtime.NewString("/users"),
	})
	if err == nil {
		t.Fatal("expected unsupported dialect error")
	}
	if !strings.Contains(err.Error(), `unsupported dialect "sql"`) {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() {
		t.Fatal("expected unsupported dialect to fail before request")
	}
}

func TestClientUsesFerretHTTPClientFromContext(t *testing.T) {
	t.Parallel()

	httpClient := &recordingHTTPClient{
		response: &ferrethttp.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Headers: ferrethttp.Headers{
				"X-Result": []string{"ok"},
			},
			Body: []byte(`{"ok":true}`),
		},
	}
	ctx := ferretnet.WithNetwork(
		context.Background(),
		ferretnet.New(ferretnet.WithHTTPClient(httpClient)),
	)

	cfg := DefaultConfig()
	cfg.BaseURL = "https://api.example.test/v1/"
	cfg.Headers.Set("Authorization", "Bearer token")
	client := NewClient(cfg)

	out, err := client.QueryOne(ctx, runtime.Query{
		Expression: runtime.NewString("/users"),
		Params: object(t, map[string]runtime.Value{
			"method": runtime.NewString("POST"),
			"query": object(t, map[string]runtime.Value{
				"active": runtime.True,
			}),
			"headers": object(t, map[string]runtime.Value{
				"X-Request-ID": runtime.NewString("req-1"),
			}),
			"body": object(t, map[string]runtime.Value{
				"name": runtime.NewString("Ada"),
			}),
		}),
		Options: object(t, map[string]runtime.Value{
			"response": runtime.NewString("full"),
		}),
	})
	if err != nil {
		t.Fatalf("unexpected query error: %v", err)
	}
	if got := field(t, out, "status"); got != runtime.NewInt(http.StatusOK) {
		t.Fatalf("expected status 200, got %s", got.String())
	}

	req := httpClient.request()
	if req == nil {
		t.Fatal("expected HTTP client to be called")
	}
	if req.Method != http.MethodPost {
		t.Fatalf("expected POST, got %s", req.Method)
	}
	if req.URL != "https://api.example.test/users?active=true" {
		t.Fatalf("unexpected URL: %s", req.URL)
	}
	if got := req.Headers["Authorization"]; len(got) != 1 || got[0] != "Bearer token" {
		t.Fatalf("unexpected authorization header: %v", got)
	}
	if got := req.Headers["X-Request-Id"]; len(got) != 1 || got[0] != "req-1" {
		t.Fatalf("unexpected request id header: %v", got)
	}
	if got := req.Headers["Content-Type"]; len(got) != 1 || got[0] != "application/json" {
		t.Fatalf("unexpected content type: %v", got)
	}
	if string(req.Body) != `{"name":"Ada"}` {
		t.Fatalf("unexpected body: %s", req.Body)
	}
}

func TestClientRequiresNetworkContext(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.BaseURL = "https://api.example.test"
	client := NewClient(cfg)

	_, err := client.QueryOne(context.Background(), runtime.Query{Expression: runtime.NewString("/users")})
	if err == nil {
		t.Fatal("expected missing network error")
	}
	if !strings.Contains(err.Error(), "network not found in context") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func object(t *testing.T, props map[string]runtime.Value) *runtime.Object {
	t.Helper()

	ctx := context.Background()
	out := runtime.NewObjectOf(len(props))
	for key, value := range props {
		if err := out.Set(ctx, runtime.NewString(key), value); err != nil {
			t.Fatalf("failed to set object property %s: %v", key, err)
		}
	}

	return out
}

func field(t *testing.T, value runtime.Value, key string) runtime.Value {
	t.Helper()

	obj, ok := value.(runtime.Map)
	if !ok {
		t.Fatalf("expected object for %s, got %T", key, value)
	}

	out, found, err := obj.Lookup(context.Background(), runtime.NewString(key))
	if err != nil {
		t.Fatalf("failed to lookup %s: %v", key, err)
	}
	if !found {
		t.Fatalf("expected field %s", key)
	}

	return out
}

func readBody(t *testing.T, r *http.Request) string {
	t.Helper()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	return string(data)
}

func networkContext() context.Context {
	return ferretnet.WithNetwork(context.Background(), ferretnet.New())
}
