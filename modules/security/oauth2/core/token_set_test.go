package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTokenSetFormattingAndSerializationRedactSecrets(t *testing.T) {
	secretValues := []string{
		"access-secret",
		"refresh-secret",
		"id-secret",
		"unknown-secret",
	}
	token := &TokenSet{
		AccessToken:  secretValues[0],
		TokenType:    "Bearer",
		RefreshToken: secretValues[1],
		Scope:        "read write",
		ExpiresAt:    time.Date(2026, 7, 24, 14, 0, 0, 0, time.UTC),
		IDToken:      secretValues[2],
		Extra: map[string]any{
			"provider_credential": secretValues[3],
		},
	}

	jsonValue, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("marshal token: %v", err)
	}

	formatted := []string{
		token.String(),
		fmt.Sprintf("%v", token),
		fmt.Sprintf("%+v", token),
		fmt.Sprintf("%#v", token),
		string(jsonValue),
	}
	for _, value := range formatted {
		for _, secret := range secretValues {
			if strings.Contains(value, secret) {
				t.Fatalf("safe representation leaked %q: %s", secret, value)
			}
		}
	}
	if !strings.Contains(token.String(), redactedValue) {
		t.Fatalf("expected redaction marker in token String: %s", token.String())
	}

	var projection map[string]any
	if err := json.Unmarshal(jsonValue, &projection); err != nil {
		t.Fatalf("decode projection: %v", err)
	}
	if projection["access_token"] != redactedValue ||
		projection["refresh_token"] != redactedValue ||
		projection["id_token"] != redactedValue {
		t.Fatalf("unexpected redacted projection: %v", projection)
	}
	if _, found := projection["Extra"]; found {
		t.Fatalf("Extra leaked into safe serialization: %v", projection)
	}
}

func TestTokenSetSerializationOmitsEmptySecrets(t *testing.T) {
	value, err := (&TokenSet{TokenType: "Bearer"}).MarshalJSON()
	if err != nil {
		t.Fatalf("marshal token: %v", err)
	}

	for _, field := range []string{"access_token", "refresh_token", "id_token"} {
		if strings.Contains(string(value), field) {
			t.Fatalf("empty secret field %q was serialized: %s", field, value)
		}
	}
}

func TestTokenSetCloneDeepCopiesExtra(t *testing.T) {
	original := &TokenSet{
		Extra: map[string]any{
			"nested": map[string]any{
				"values": []any{"one", map[string]any{"key": "value"}},
			},
		},
	}
	cloned := original.Clone()

	nested := cloned.Extra["nested"].(map[string]any)
	values := nested["values"].([]any)
	values[0] = "changed"
	values[1].(map[string]any)["key"] = "changed"

	originalNested := original.Extra["nested"].(map[string]any)
	originalValues := originalNested["values"].([]any)
	if originalValues[0] != "one" ||
		originalValues[1].(map[string]any)["key"] != "value" {
		t.Fatalf("Clone shared nested Extra state: %#v", original.Extra)
	}
}

func TestTokenSetExpiryBoundaries(t *testing.T) {
	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)

	unknown := TokenSet{}
	if unknown.Expired(now, 0) {
		t.Fatal("unknown expiry was considered expired")
	}
	if _, known := unknown.ValidFor(now); known {
		t.Fatal("unknown expiry reported a lifetime")
	}

	token := TokenSet{ExpiresAt: now.Add(1500 * time.Millisecond)}
	if token.Expired(now, 1499*time.Millisecond) {
		t.Fatal("token expired before skew boundary")
	}
	if !token.Expired(now, 1500*time.Millisecond) {
		t.Fatal("token did not expire at skew boundary")
	}
	if remaining, known := token.ValidFor(now); !known ||
		remaining != 1500*time.Millisecond {
		t.Fatalf("unexpected lifetime: %v, %t", remaining, known)
	}
	if remaining, known := token.ValidFor(now.Add(2 * time.Second)); !known || remaining != 0 {
		t.Fatalf("expired lifetime was not clamped: %v, %t", remaining, known)
	}
}

func TestTokenSetScopesUsesASCIISpaces(t *testing.T) {
	token := TokenSet{Scope: "read  write "}
	if got := token.Scopes(); fmt.Sprint(got) != "[read write]" {
		t.Fatalf("unexpected scopes: %v", got)
	}
}
