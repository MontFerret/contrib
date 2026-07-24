package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type discoveryHTTPClient struct {
	request  *ferrethttp.Request
	response *ferrethttp.Response
	err      error
}

func (c *discoveryHTTPClient) Do(_ context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
	c.request = request

	return c.response, c.err
}

func TestDiscoverLoadsMetadataAndAppliesRFCDefaults(t *testing.T) {
	t.Parallel()

	httpClient := &discoveryHTTPClient{
		response: &ferrethttp.Response{
			StatusCode: 200,
			Headers: ferrethttp.Headers{
				"content-type": {"application/json; charset=utf-8"},
			},
			Body: []byte(`{
				"issuer": "https://auth.example.com/tenant",
				"authorization_endpoint": "https://auth.example.com/authorize",
				"token_endpoint": "https://auth.example.com/token",
				"scopes_supported": ["users:read"],
				"extension_metadata": {"ignored": true}
			}`),
		},
	}

	provider, err := Discover(
		context.Background(),
		httpClient,
		"https://auth.example.com/tenant",
		DiscoveryOptions{},
	)
	if err != nil {
		t.Fatalf("unexpected discovery error: %v", err)
	}

	if got := httpClient.request.URL; got != "https://auth.example.com/.well-known/oauth-authorization-server/tenant" {
		t.Fatalf("unexpected well-known URL: %q", got)
	}
	if got := httpClient.request.Method; got != "GET" {
		t.Fatalf("unexpected discovery method: %q", got)
	}
	if got := httpClient.request.Headers["Accept"]; len(got) != 1 || got[0] != "application/json" {
		t.Fatalf("unexpected Accept header: %#v", got)
	}
	if got := provider.TokenEndpointAuthMethods; len(got) != 1 || got[0] != string(ClientAuthMethodBasic) {
		t.Fatalf("unexpected RFC auth-method default: %#v", got)
	}
	if got := provider.GrantTypesSupported; len(got) != 2 ||
		got[0] != "authorization_code" ||
		got[1] != "implicit" {
		t.Fatalf("unexpected RFC grant default: %#v", got)
	}

	httpClient.response.Body[0] = ' '
	if provider.Issuer != "https://auth.example.com/tenant" {
		t.Fatal("provider unexpectedly retained response storage")
	}
}

func TestDiscoverPreservesExplicitMetadata(t *testing.T) {
	t.Parallel()

	httpClient := &discoveryHTTPClient{
		response: &ferrethttp.Response{
			StatusCode: 200,
			Headers: ferrethttp.Headers{
				"Content-Type": {"application/json"},
			},
			Body: []byte(`{
				"issuer": "http://127.0.0.1:8080",
				"token_endpoint": "http://127.0.0.1:8080/token",
				"grant_types_supported": ["client_credentials"],
				"token_endpoint_auth_methods_supported": ["client_secret_post"]
			}`),
		},
	}

	provider, err := Discover(
		context.Background(),
		httpClient,
		"http://127.0.0.1:8080",
		DiscoveryOptions{InsecureAllowHTTP: true},
	)
	if err != nil {
		t.Fatalf("unexpected development discovery error: %v", err)
	}

	if got := provider.GrantTypesSupported; len(got) != 1 || got[0] != "client_credentials" {
		t.Fatalf("unexpected explicit grants: %#v", got)
	}
	if got := provider.TokenEndpointAuthMethods; len(got) != 1 ||
		got[0] != string(ClientAuthMethodPost) {
		t.Fatalf("unexpected explicit auth methods: %#v", got)
	}
}

func TestDiscoverRejectsInvalidResponses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response *ferrethttp.Response
		match    string
	}{
		{
			name: "non-200 status",
			response: &ferrethttp.Response{
				StatusCode: 404,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
			},
			match: "HTTP status 404",
		},
		{
			name: "missing content type",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Body:       []byte(`{}`),
			},
			match: "Content-Type must be application/json",
		},
		{
			name: "unsupported content type",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"text/plain"},
				},
				Body: []byte(`{}`),
			},
			match: "must be application/json",
		},
		{
			name: "malformed JSON",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"issuer":`),
			},
			match: "not valid JSON",
		},
		{
			name: "non-object JSON",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`null`),
			},
			match: "JSON object",
		},
		{
			name: "missing issuer",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"token_endpoint":"https://auth.example.com/token"}`),
			},
			match: "issuer is required",
		},
		{
			name: "issuer mismatch is exact",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{
					"issuer":"https://auth.example.com/tenant/",
					"token_endpoint":"https://auth.example.com/token"
				}`),
			},
			match: "does not exactly match",
		},
		{
			name: "missing token endpoint",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{"issuer":"https://auth.example.com/tenant"}`),
			},
			match: "token_endpoint is required",
		},
		{
			name: "invalid token endpoint",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{
					"issuer":"https://auth.example.com/tenant",
					"token_endpoint":"/token"
				}`),
			},
			match: "absolute HTTP(S)",
		},
		{
			name: "malformed known metadata",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers: ferrethttp.Headers{
					"Content-Type": {"application/json"},
				},
				Body: []byte(`{
					"issuer":"https://auth.example.com/tenant",
					"token_endpoint":"https://auth.example.com/token",
					"scopes_supported":"users:read"
				}`),
			},
			match: "array of strings",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			httpClient := &discoveryHTTPClient{response: test.response}
			_, err := Discover(
				context.Background(),
				httpClient,
				"https://auth.example.com/tenant",
				DiscoveryOptions{},
			)
			if err == nil {
				t.Fatal("expected discovery error")
			}
			if !strings.Contains(err.Error(), test.match) {
				t.Fatalf("expected error to contain %q, got %v", test.match, err)
			}
		})
	}
}

func TestDiscoverPreservesTransportErrorsAndValidatesOptions(t *testing.T) {
	t.Parallel()

	transportErr := errors.New("policy denied localhost")
	_, err := Discover(
		context.Background(),
		&discoveryHTTPClient{err: transportErr},
		"https://auth.example.com",
		DiscoveryOptions{},
	)
	if !errors.Is(err, transportErr) {
		t.Fatalf("transport error was not preserved: %v", err)
	}

	_, err = Discover(
		context.Background(),
		&discoveryHTTPClient{},
		"https://auth.example.com",
		DiscoveryOptions{Timeout: -time.Millisecond},
	)
	if err == nil || !strings.Contains(err.Error(), "must not be negative") {
		t.Fatalf("expected negative timeout error, got %v", err)
	}

	_, err = Discover(
		context.Background(),
		&discoveryHTTPClient{},
		"http://127.0.0.1:8080",
		DiscoveryOptions{},
	)
	if err == nil || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("expected default HTTP rejection, got %v", err)
	}
}

func TestDiscoverRetainsFerretNetworkPolicy(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/.well-known/oauth-authorization-server" {
			http.NotFound(w, request)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":         server.URL,
			"token_endpoint": server.URL + "/token",
		})
	}))
	defer server.Close()

	deniedClient, err := ferrethttp.New()
	if err != nil {
		t.Fatalf("failed to create default HTTP client: %v", err)
	}

	_, err = Discover(
		context.Background(),
		deniedClient,
		server.URL,
		DiscoveryOptions{InsecureAllowHTTP: true},
	)
	if !errors.Is(err, ferrethttp.ErrPolicyDenied) {
		t.Fatalf("expected localhost policy denial, got %v", err)
	}

	allowedClient, err := ferrethttp.New(ferrethttp.WithAllowLocalhost(true))
	if err != nil {
		t.Fatalf("failed to create localhost HTTP client: %v", err)
	}

	provider, err := Discover(
		context.Background(),
		allowedClient,
		server.URL,
		DiscoveryOptions{InsecureAllowHTTP: true},
	)
	if err != nil {
		t.Fatalf("unexpected explicitly allowed discovery error: %v", err)
	}
	if provider.TokenEndpoint != server.URL+"/token" {
		t.Fatalf("unexpected discovered token endpoint: %q", provider.TokenEndpoint)
	}
}

func TestWellKnownURLInsertsBeforeIssuerPath(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"https://auth.example.com":         "https://auth.example.com/.well-known/oauth-authorization-server",
		"https://auth.example.com/":        "https://auth.example.com/.well-known/oauth-authorization-server/",
		"https://auth.example.com/a/b":     "https://auth.example.com/.well-known/oauth-authorization-server/a/b",
		"https://auth.example.com/a%2Fb":   "https://auth.example.com/.well-known/oauth-authorization-server/a%2Fb",
		"https://auth.example.com/tenant/": "https://auth.example.com/.well-known/oauth-authorization-server/tenant/",
	}

	for issuer, expected := range tests {
		issuer := issuer
		expected := expected
		t.Run(issuer, func(t *testing.T) {
			t.Parallel()

			got, err := wellKnownURL(issuer, false)
			if err != nil {
				t.Fatalf("unexpected well-known URL error: %v", err)
			}
			if got != expected {
				t.Fatalf("expected %q, got %q", expected, got)
			}
		})
	}
}
