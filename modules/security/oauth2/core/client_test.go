package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNewClientSelectsAuthenticationDefaults(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t, nil)

	confidential, err := NewClient(provider, ClientConfig{
		ClientID:     "confidential",
		ClientSecret: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected confidential client error: %v", err)
	}
	if confidential.AuthMethod != ClientAuthMethodBasic {
		t.Fatalf("expected Basic default, got %q", confidential.AuthMethod)
	}

	public, err := NewClient(provider, ClientConfig{ClientID: "public"})
	if err != nil {
		t.Fatalf("unexpected public client error: %v", err)
	}
	if public.AuthMethod != ClientAuthMethodNone {
		t.Fatalf("expected none default, got %q", public.AuthMethod)
	}
}

func TestNewClientValidatesAuthentication(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t, nil)

	tests := []struct {
		name   string
		config ClientConfig
		match  string
	}{
		{
			name:   "missing client ID",
			config: ClientConfig{},
			match:  "client ID is required",
		},
		{
			name: "Basic requires secret",
			config: ClientConfig{
				ClientID:   "client",
				AuthMethod: ClientAuthMethodBasic,
			},
			match: "requires a client secret",
		},
		{
			name: "post requires secret",
			config: ClientConfig{
				ClientID:   "client",
				AuthMethod: ClientAuthMethodPost,
			},
			match: "requires a client secret",
		},
		{
			name: "none rejects secret",
			config: ClientConfig{
				ClientID:     "client",
				ClientSecret: "secret",
				AuthMethod:   ClientAuthMethodNone,
			},
			match: "does not permit a client secret",
		},
		{
			name: "unsupported method",
			config: ClientConfig{
				ClientID:   "client",
				AuthMethod: "private_key_jwt",
			},
			match: "unsupported",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewClient(provider, test.config)
			if err == nil {
				t.Fatal("expected client error")
			}
			if !strings.Contains(err.Error(), test.match) {
				t.Fatalf("expected error to contain %q, got %v", test.match, err)
			}
		})
	}

	for _, method := range []ClientAuthMethod{ClientAuthMethodBasic, ClientAuthMethodPost} {
		client, err := NewClient(provider, ClientConfig{
			ClientID:     "client",
			ClientSecret: "secret",
			AuthMethod:   method,
		})
		if err != nil {
			t.Fatalf("unexpected %q client error: %v", method, err)
		}
		if client.AuthMethod != method {
			t.Fatalf("expected method %q, got %q", method, client.AuthMethod)
		}
	}
}

func TestNewClientValidatesAdvertisedAuthenticationMethods(t *testing.T) {
	t.Parallel()

	unknownMetadata := mustProvider(t, nil)
	if _, err := NewClient(unknownMetadata, ClientConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		AuthMethod:   ClientAuthMethodPost,
	}); err != nil {
		t.Fatalf("omitted metadata should not restrict the method: %v", err)
	}

	postOnly := mustProvider(t, []string{string(ClientAuthMethodPost)})
	if _, err := NewClient(postOnly, ClientConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		AuthMethod:   ClientAuthMethodBasic,
	}); err == nil || !strings.Contains(err.Error(), "does not advertise") {
		t.Fatalf("expected advertised method mismatch, got %v", err)
	}

	explicitEmpty := mustProvider(t, []string{})
	if _, err := NewClient(explicitEmpty, ClientConfig{ClientID: "client"}); err == nil {
		t.Fatal("explicitly empty advertised methods should reject all methods")
	}
}

func TestClientCopiesProviderAndRevalidatesMutableState(t *testing.T) {
	t.Parallel()

	provider := mustProvider(t, nil)
	client, err := NewClient(provider, ClientConfig{
		ClientID:     "client",
		ClientSecret: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	provider.TokenEndpoint = "https://changed.example.com/token"
	if got := client.Provider.TokenEndpoint; got != "https://auth.example.com/token" {
		t.Fatalf("client retained provider pointer: %q", got)
	}

	clone := client.Clone()
	clone.Provider.TokenEndpoint = "https://clone.example.com/token"
	if got := client.Provider.TokenEndpoint; got != "https://auth.example.com/token" {
		t.Fatalf("client clone retained provider pointer: %q", got)
	}

	client.Provider.TokenEndpoint = "http://auth.example.com/token"
	if err := client.Validate(); err == nil || !strings.Contains(err.Error(), "provider is invalid") {
		t.Fatalf("expected mutated provider validation error, got %v", err)
	}

	client.Provider.TokenEndpoint = "https://auth.example.com/token"
	client.AuthMethod = ClientAuthMethodNone
	if err := client.Validate(); err == nil || !strings.Contains(err.Error(), "does not permit") {
		t.Fatalf("expected mutated auth validation error, got %v", err)
	}
}

func TestClientFormattingAndJSONRedactSecret(t *testing.T) {
	t.Parallel()

	const secret = "super-secret-client-value"

	client, err := NewClient(mustProvider(t, nil), ClientConfig{
		ClientID:     "client",
		ClientSecret: secret,
	})
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}

	formatted := []string{
		fmt.Sprintf("%v", client),
		fmt.Sprintf("%+v", client),
		fmt.Sprintf("%#v", client),
		client.String(),
	}
	for _, value := range formatted {
		if strings.Contains(value, secret) {
			t.Fatalf("client secret leaked through formatting: %s", value)
		}
		if !strings.Contains(value, "<redacted>") {
			t.Fatalf("client secret was not visibly redacted: %s", value)
		}
	}

	data, err := json.Marshal(client)
	if err != nil {
		t.Fatalf("unexpected JSON error: %v", err)
	}
	if strings.Contains(string(data), secret) {
		t.Fatalf("client secret leaked through JSON: %s", data)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode client JSON: %v", err)
	}
	if decoded["client_secret"] != "<redacted>" {
		t.Fatalf("client JSON did not redact the secret: %s", data)
	}
}

func mustProvider(t *testing.T, authMethods []string) *Provider {
	t.Helper()

	provider, err := NewProvider(ProviderConfig{
		TokenEndpoint:            "https://auth.example.com/token",
		TokenEndpointAuthMethods: authMethods,
	})
	if err != nil {
		t.Fatalf("failed to build test provider: %v", err)
	}

	return provider
}
