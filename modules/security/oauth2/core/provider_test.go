package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNewProviderValidatesAndCopiesConfiguration(t *testing.T) {
	t.Parallel()

	scopes := []string{"users:read"}
	grants := []string{"client_credentials"}
	authMethods := []string{string(ClientAuthMethodBasic)}

	provider, err := NewProvider(ProviderConfig{
		Issuer:                   "https://auth.example.com/tenant",
		AuthorizationEndpoint:    "https://auth.example.com/authorize",
		TokenEndpoint:            "https://auth.example.com/token?resource=api",
		RevocationEndpoint:       "https://auth.example.com/revoke",
		IntrospectionEndpoint:    "https://auth.example.com/introspect",
		JWKSURI:                  "https://auth.example.com/jwks.json",
		ScopesSupported:          scopes,
		GrantTypesSupported:      grants,
		TokenEndpointAuthMethods: authMethods,
	})
	if err != nil {
		t.Fatalf("unexpected provider error: %v", err)
	}

	scopes[0] = "changed"
	grants[0] = "changed"
	authMethods[0] = string(ClientAuthMethodPost)

	if got := provider.ScopesSupported[0]; got != "users:read" {
		t.Fatalf("constructor retained scopes slice: %q", got)
	}
	if got := provider.GrantTypesSupported[0]; got != "client_credentials" {
		t.Fatalf("constructor retained grants slice: %q", got)
	}
	if got := provider.TokenEndpointAuthMethods[0]; got != string(ClientAuthMethodBasic) {
		t.Fatalf("constructor retained auth methods slice: %q", got)
	}

	clone := provider.Clone()
	clone.ScopesSupported[0] = "clone-change"
	if got := provider.ScopesSupported[0]; got != "users:read" {
		t.Fatalf("clone retained scopes slice: %q", got)
	}

	provider.TokenEndpoint = "http://auth.example.com/token"
	if err := provider.Validate(); err == nil || !strings.Contains(err.Error(), "must use HTTPS") {
		t.Fatalf("expected mutable configuration to be revalidated, got %v", err)
	}
}

func TestNewProviderAllowsExplicitDevelopmentHTTP(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(ProviderConfig{
		Issuer:            "http://127.0.0.1:8080",
		TokenEndpoint:     "http://127.0.0.1:8080/token",
		InsecureAllowHTTP: true,
	})
	if err != nil {
		t.Fatalf("unexpected HTTP provider error: %v", err)
	}

	if err := provider.Validate(); err != nil {
		t.Fatalf("unexpected revalidation error: %v", err)
	}
}

func TestNewProviderRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		match  string
		config ProviderConfig
	}{
		{
			name:   "missing token endpoint",
			config: ProviderConfig{},
			match:  "token endpoint is required",
		},
		{
			name: "relative token endpoint",
			config: ProviderConfig{
				TokenEndpoint: "/token",
			},
			match: "absolute HTTP(S)",
		},
		{
			name: "unsupported token endpoint scheme",
			config: ProviderConfig{
				TokenEndpoint: "ftp://auth.example.com/token",
			},
			match: "HTTP or HTTPS",
		},
		{
			name: "HTTP rejected by default",
			config: ProviderConfig{
				TokenEndpoint: "http://auth.example.com/token",
			},
			match: "must use HTTPS",
		},
		{
			name: "userinfo rejected",
			config: ProviderConfig{
				TokenEndpoint: "https://client:secret@auth.example.com/token",
			},
			match: "user information",
		},
		{
			name: "fragment rejected",
			config: ProviderConfig{
				TokenEndpoint: "https://auth.example.com/token#fragment",
			},
			match: "fragment",
		},
		{
			name: "issuer query rejected",
			config: ProviderConfig{
				Issuer:        "https://auth.example.com?tenant=one",
				TokenEndpoint: "https://auth.example.com/token",
			},
			match: "issuer must not contain a query",
		},
		{
			name: "sensitive token query rejected",
			config: ProviderConfig{
				TokenEndpoint: "https://auth.example.com/token?CLIENT_SECRET=secret",
			},
			match: "sensitive query parameter",
		},
		{
			name: "client ID token query rejected",
			config: ProviderConfig{
				TokenEndpoint: "https://auth.example.com/token?client_id=client",
			},
			match: "sensitive query parameter",
		},
		{
			name: "encoded sensitive token query rejected",
			config: ProviderConfig{
				TokenEndpoint: "https://auth.example.com/token?access%5Ftoken=secret",
			},
			match: "sensitive query parameter",
		},
		{
			name: "invalid optional endpoint rejected",
			config: ProviderConfig{
				TokenEndpoint:         "https://auth.example.com/token",
				RevocationEndpoint:    "https://auth.example.com/revoke",
				IntrospectionEndpoint: "relative",
			},
			match: "introspection endpoint",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewProvider(test.config)
			if err == nil {
				t.Fatal("expected provider error")
			}
			if !strings.Contains(err.Error(), test.match) {
				t.Fatalf("expected error to contain %q, got %v", test.match, err)
			}
		})
	}
}

func TestProviderFormattingAndJSONAreSafe(t *testing.T) {
	t.Parallel()

	provider, err := NewProvider(ProviderConfig{
		TokenEndpoint: "https://auth.example.com/token",
	})
	if err != nil {
		t.Fatalf("unexpected provider error: %v", err)
	}

	for _, formatted := range []string{
		fmt.Sprintf("%v", provider),
		fmt.Sprintf("%+v", provider),
		fmt.Sprintf("%#v", provider),
		provider.String(),
	} {
		if !strings.Contains(formatted, `"type":"oauth2.Provider"`) {
			t.Fatalf("unexpected provider formatting: %s", formatted)
		}
	}

	data, err := json.Marshal(provider)
	if err != nil {
		t.Fatalf("unexpected JSON error: %v", err)
	}
	if !strings.Contains(string(data), `"token_endpoint":"https://auth.example.com/token"`) {
		t.Fatalf("unexpected provider JSON: %s", data)
	}
	if strings.Contains(string(data), "insecureAllowHTTP") {
		t.Fatalf("private validation state leaked in JSON: %s", data)
	}
}
