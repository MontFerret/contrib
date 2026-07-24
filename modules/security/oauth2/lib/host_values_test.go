package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	encodingjson "github.com/MontFerret/ferret/v2/pkg/encoding/json"
	encodingmsgpack "github.com/MontFerret/ferret/v2/pkg/encoding/msgpack"
	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/MontFerret/contrib/modules/security/oauth2/core"
)

func TestHostValuesKeepSecretsOutOfFormattingAndSerialization(t *testing.T) {
	t.Parallel()

	const (
		clientSecret = "client-secret-value"
		accessToken  = "access-token-value"
		refreshToken = "refresh-token-value"
		idToken      = "id-token-value"
		extraSecret  = "provider-extra-secret"
	)

	provider, err := core.NewProvider(core.ProviderConfig{
		Issuer:        "https://issuer.example",
		TokenEndpoint: "https://issuer.example/token",
	})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	client, err := core.NewClient(provider, core.ClientConfig{
		ClientID:     "client-id",
		ClientSecret: clientSecret,
	})
	if err != nil {
		t.Fatalf("construct client: %v", err)
	}
	token := &core.TokenSet{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		IDToken:      idToken,
		Scope:        "users:read",
		Extra: map[string]any{
			"secret": extraSecret,
		},
	}

	values := []runtime.Value{
		newProviderValue(provider),
		newClientValue(client),
		newTokenValue(token),
	}
	secrets := []string{clientSecret, accessToken, refreshToken, idToken, extraSecret}

	for _, value := range values {
		assertNoSecrets(t, []byte(fmt.Sprintf("%v", value)), secrets)
		assertNoSecrets(t, []byte(fmt.Sprintf("%+v", value)), secrets)
		assertNoSecrets(t, []byte(fmt.Sprintf("%#v", value)), secrets)

		goJSON, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			t.Fatalf("Go JSON encode %T: %v", value, marshalErr)
		}
		assertNoSecrets(t, goJSON, secrets)

		ferretJSON, encodeErr := encodingjson.Default.Encode(value)
		if encodeErr != nil {
			t.Fatalf("Ferret JSON encode %T: %v", value, encodeErr)
		}
		assertNoSecrets(t, ferretJSON, secrets)

		messagePack, encodeErr := encodingmsgpack.Default.Encode(value)
		if encodeErr != nil {
			t.Fatalf("MessagePack encode %T: %v", value, encodeErr)
		}
		assertNoSecrets(t, messagePack, secrets)
	}
}

func TestProviderAndClientHostPropertiesAreSafe(t *testing.T) {
	t.Parallel()

	provider, err := core.NewProvider(core.ProviderConfig{
		Issuer:                   "https://issuer.example",
		TokenEndpoint:            "https://issuer.example/token",
		ScopesSupported:          []string{"read"},
		TokenEndpointAuthMethods: []string{"client_secret_basic"},
	})
	if err != nil {
		t.Fatalf("construct provider: %v", err)
	}
	client, err := core.NewClient(provider, core.ClientConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("construct client: %v", err)
	}

	ctx := context.Background()
	providerHost := newProviderValue(provider)
	clientHost := newClientValue(client)

	assertPropertyString(t, ctx, providerHost, "issuer", "https://issuer.example")
	assertPropertyString(
		t,
		ctx,
		providerHost,
		"tokenEndpoint",
		"https://issuer.example/token",
	)
	assertPropertyString(t, ctx, clientHost, "clientID", "client-id")
	assertPropertyString(t, ctx, clientHost, "authMethod", "client_secret_basic")
	assertPropertyBool(t, ctx, clientHost, "hasClientSecret", true)

	secret, err := lookupValue(ctx, clientHost, "clientSecret")
	if err != nil {
		t.Fatalf("read clientSecret: %v", err)
	}
	if secret != runtime.None {
		t.Fatalf("clientSecret property = %v, want NONE", secret)
	}
}

func TestTokenHostPropertiesAndExplicitAccessors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	token := newTokenValueWithClock(&core.TokenSet{
		AccessToken:  "access",
		TokenType:    "bEaReR",
		RefreshToken: "refresh",
		IDToken:      "id",
		Scope:        "users:read  projects:read",
		ExpiresAt:    now.Add(1500 * time.Microsecond),
	}, func() time.Time {
		return now
	})

	ctx := context.Background()
	assertPropertyString(t, ctx, token, "tokenType", "bEaReR")
	assertPropertyBool(t, ctx, token, "expired", false)
	assertPropertyBool(t, ctx, token, "hasRefreshToken", true)
	assertPropertyBool(t, ctx, token, "hasIDToken", true)

	validFor, err := lookupValue(ctx, token, "validFor")
	if err != nil {
		t.Fatalf("read validFor: %v", err)
	}
	if validFor != runtime.NewInt(2) {
		t.Fatalf("validFor = %v, want 2", validFor)
	}

	scopes, err := lookupValue(ctx, token, "scopes")
	if err != nil {
		t.Fatalf("read scopes: %v", err)
	}
	expectedScopes := runtime.NewArrayWith(
		runtime.NewString("users:read"),
		runtime.NewString("projects:read"),
	)
	if scopes.String() != expectedScopes.String() {
		t.Fatalf("scopes = %s, want %s", scopes, expectedScopes)
	}

	for _, secretProperty := range []string{
		"accessToken",
		"refreshToken",
		"idToken",
		"extra",
	} {
		value, getErr := lookupValue(ctx, token, secretProperty)
		if getErr != nil {
			t.Fatalf("read %s: %v", secretProperty, getErr)
		}
		if value != runtime.None {
			t.Fatalf("%s property = %v, want NONE", secretProperty, value)
		}
	}

	assertFunctionString(t, AccessToken, token, "access")
	assertFunctionString(t, RefreshToken, token, "refresh")
	assertFunctionString(t, IDToken, token, "id")

	header, err := AuthHeader(ctx, token)
	if err != nil {
		t.Fatalf("AUTH_HEADER: %v", err)
	}
	authorization, err := header.(runtime.KeyReadable).Get(ctx, runtime.NewString("Authorization"))
	if err != nil {
		t.Fatalf("read Authorization: %v", err)
	}
	if authorization.String() != "Bearer access" {
		t.Fatalf("Authorization = %q", authorization)
	}
}

func TestTokenHelpersReturnNoneForUnknownOptionalValues(t *testing.T) {
	t.Parallel()

	token := newTokenValue(&core.TokenSet{
		AccessToken: "access",
		TokenType:   "Bearer",
	})

	for name, function := range map[string]runtime.Function{
		"REFRESH_TOKEN": RefreshToken,
		"ID_TOKEN":      IDToken,
		"EXPIRES_AT":    ExpiresAt,
		"VALID_FOR":     ValidFor,
	} {
		value, err := function(context.Background(), token)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if value != runtime.None {
			t.Fatalf("%s = %v, want NONE", name, value)
		}
	}
}

func TestExpiredUsesIntegerMillisecondSkew(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	token := newTokenValueWithClock(&core.TokenSet{
		AccessToken: "access",
		TokenType:   "Bearer",
		ExpiresAt:   now.Add(30 * time.Millisecond),
	}, func() time.Time {
		return now
	})

	expired, err := Expired(
		context.Background(),
		token,
		runtime.NewObjectWith(map[string]runtime.Value{
			"skew": runtime.NewInt(30),
		}),
	)
	if err != nil {
		t.Fatalf("EXPIRED: %v", err)
	}
	if expired != runtime.True {
		t.Fatalf("EXPIRED = %v, want true", expired)
	}

	_, err = Expired(
		context.Background(),
		token,
		runtime.NewObjectWith(map[string]runtime.Value{
			"skew": runtime.NewFloat(30),
		}),
	)
	if err == nil {
		t.Fatal("EXPIRED accepted a floating-point duration")
	}
}

func TestAuthHeaderRejectsNonBearerAndInjection(t *testing.T) {
	t.Parallel()

	tests := []*core.TokenSet{
		{AccessToken: "access", TokenType: "MAC"},
		{AccessToken: "access\r\nInjected: true", TokenType: "Bearer"},
		{AccessToken: "", TokenType: "Bearer"},
	}

	for _, token := range tests {
		if _, err := AuthHeader(context.Background(), newTokenValue(token)); err == nil {
			t.Fatalf("expected AUTH_HEADER error for %#v", token)
		}
	}
}

func assertNoSecrets(t testing.TB, data []byte, secrets []string) {
	t.Helper()

	for _, secret := range secrets {
		if bytes.Contains(data, []byte(secret)) {
			t.Fatalf("serialized value contains secret %q: %q", secret, data)
		}
	}
}

func assertPropertyString(
	t testing.TB,
	ctx context.Context,
	value runtime.KeyReadable,
	key string,
	expected string,
) {
	t.Helper()

	actual, err := lookupValue(ctx, value, key)
	if err != nil {
		t.Fatalf("read %s: %v", key, err)
	}
	if actual.String() != expected {
		t.Fatalf("%s = %q, want %q", key, actual, expected)
	}
}

func assertPropertyBool(
	t testing.TB,
	ctx context.Context,
	value runtime.KeyReadable,
	key string,
	expected bool,
) {
	t.Helper()

	actual, err := lookupValue(ctx, value, key)
	if err != nil {
		t.Fatalf("read %s: %v", key, err)
	}
	if actual != runtime.NewBoolean(expected) {
		t.Fatalf("%s = %v, want %v", key, actual, expected)
	}
}

func assertFunctionString(
	t testing.TB,
	function runtime.Function,
	token runtime.Value,
	expected string,
) {
	t.Helper()

	actual, err := function(context.Background(), token)
	if err != nil {
		t.Fatalf("call function: %v", err)
	}
	if actual.String() != expected {
		t.Fatalf("result = %q, want %q", actual, expected)
	}
}
