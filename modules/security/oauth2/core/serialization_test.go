package core

import (
	"bytes"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestCoreMessagePackSerializationRedactsSecrets(t *testing.T) {
	t.Parallel()

	const (
		clientSecret = "client-secret-messagepack"
		accessToken  = "access-token-messagepack"
		refreshToken = "refresh-token-messagepack"
		idToken      = "id-token-messagepack"
		extraSecret  = "extra-secret-messagepack"
	)

	provider, err := NewProvider(ProviderConfig{
		TokenEndpoint: "https://auth.example.com/token",
	})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	client, err := NewClient(provider, ClientConfig{
		ClientID:     "client",
		ClientSecret: clientSecret,
	})
	if err != nil {
		t.Fatalf("construct client: %v", err)
	}
	token := &TokenSet{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Extra: map[string]any{
			"credential": extraSecret,
		},
	}

	for _, value := range []any{provider, client, token} {
		data, err := msgpack.Marshal(value)
		if err != nil {
			t.Fatalf("marshal %T: %v", value, err)
		}

		for _, secret := range []string{
			clientSecret,
			accessToken,
			refreshToken,
			idToken,
			extraSecret,
		} {
			if bytes.Contains(data, []byte(secret)) {
				t.Fatalf("MessagePack for %T leaked %q", value, secret)
			}
		}
	}
}
