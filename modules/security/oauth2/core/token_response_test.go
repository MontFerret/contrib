package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func TestParseTokenSuccessRequiresKnownFieldsAndRejectsDuplicates(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing access token",
			body: `{"token_type":"Bearer"}`,
		},
		{
			name: "empty access token",
			body: `{"access_token":"","token_type":"Bearer"}`,
		},
		{
			name: "missing token type",
			body: `{"access_token":"access"}`,
		},
		{
			name: "empty token type",
			body: `{"access_token":"access","token_type":""}`,
		},
		{
			name: "duplicate known field",
			body: `{"access_token":"one","access_token":"two","token_type":"Bearer"}`,
		},
		{
			name: "wrong known field type",
			body: `{"access_token":"access","token_type":"Bearer","scope":["read"]}`,
		},
		{
			name: "malformed scope",
			body: `{"access_token":"access","token_type":"Bearer","scope":"read  write"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseTokenSuccess([]byte(test.body), time.Now())
			if !errors.Is(err, ErrInvalidTokenResponse) {
				t.Fatalf("expected ErrInvalidTokenResponse, got %v", err)
			}
		})
	}
}

func TestParseTokenSuccessStrictExpiresIn(t *testing.T) {
	invalid := []string{
		`-1`,
		`1.5`,
		`1e3`,
		`"1"`,
		`null`,
		`9223372036854775807`,
	}

	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			body := `{"access_token":"access","token_type":"Bearer","expires_in":` +
				value + `}`

			_, err := parseTokenSuccess([]byte(body), time.Now())
			if !errors.Is(err, ErrInvalidTokenResponse) {
				t.Fatalf("expected strict expires_in error, got %v", err)
			}
		})
	}

	now := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	parsed, err := parseTokenSuccess(
		[]byte(`{"access_token":"access","token_type":"Bearer","expires_in":0}`),
		now,
	)
	if err != nil {
		t.Fatalf("parse zero expires_in: %v", err)
	}
	if !parsed.token.ExpiresAt.Equal(now) {
		t.Fatalf("zero expires_in did not produce a known immediate expiry: %v", parsed.token.ExpiresAt)
	}
}

func TestParseTokenSuccessPreservesUnknownJSONNumbers(t *testing.T) {
	parsed, err := parseTokenSuccess([]byte(`{
		"access_token":"access",
		"token_type":"Bearer",
		"integer":9007199254740993,
		"decimal":1.25,
		"nested":{"value":2.50}
	}`), time.Now())
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if value, ok := parsed.token.Extra["integer"].(json.Number); !ok ||
		value.String() != "9007199254740993" {
		t.Fatalf("integer lost json.Number fidelity: %#v", parsed.token.Extra["integer"])
	}
	if value, ok := parsed.token.Extra["decimal"].(json.Number); !ok ||
		value.String() != "1.25" {
		t.Fatalf("decimal lost json.Number fidelity: %#v", parsed.token.Extra["decimal"])
	}
	nested, ok := parsed.token.Extra["nested"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected nested value: %#v", parsed.token.Extra["nested"])
	}
	if value, ok := nested["value"].(json.Number); !ok || value.String() != "2.50" {
		t.Fatalf("nested number lost fidelity: %#v", nested["value"])
	}
}

func TestParseTokenEndpointResponseRequiresExactSuccessStatusAndJSON(t *testing.T) {
	tests := []struct {
		response *ferrethttp.Response
		name     string
	}{
		{
			name: "non JSON",
			response: &ferrethttp.Response{
				StatusCode: 200,
				Headers:    ferrethttp.Headers{"Content-Type": {"text/plain"}},
				Body:       []byte(`{"access_token":"access","token_type":"Bearer"}`),
			},
		},
		{
			name: "created is not success",
			response: &ferrethttp.Response{
				StatusCode: 201,
				Headers:    ferrethttp.Headers{"Content-Type": {"application/json"}},
				Body:       []byte(`{"access_token":"access","token_type":"Bearer"}`),
			},
		},
		{
			name:     "nil response",
			response: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseTokenEndpointResponse(
				test.response,
				time.Now(),
				"",
				"",
				nil,
			)
			if !errors.Is(err, ErrInvalidTokenResponse) {
				t.Fatalf("expected ErrInvalidTokenResponse, got %v", err)
			}
		})
	}
}

func TestParseTokenOAuthErrorRejectsDuplicateKnownFields(t *testing.T) {
	_, err := parseTokenOAuthError(
		[]byte(`{"error":"one","error":"two"}`),
		400,
		nil,
	)
	if !errors.Is(err, ErrInvalidTokenResponse) {
		t.Fatalf("expected duplicate-field rejection, got %v", err)
	}
}

func TestParseTokenOAuthErrorRejectsMalformedFields(t *testing.T) {
	t.Parallel()

	tests := []string{
		`{"error":"invalid\ngrant"}`,
		`{"error":"invalid_grant","error_description":""}`,
		`{"error":"invalid_grant","error_description":"bad \"grant\""}`,
		`{"error":"invalid_grant","error_uri":"https://example.com/error with space"}`,
		`{"error":"invalid_grant","error_uri":"https://example.com/%zz"}`,
	}

	for _, body := range tests {
		_, err := parseTokenOAuthError([]byte(body), 400, nil)
		if !errors.Is(err, ErrInvalidTokenResponse) {
			t.Fatalf("expected malformed OAuth error rejection for %s, got %v", body, err)
		}
	}
}

func TestMalformedOAuthFailureIncludesHTTPStatus(t *testing.T) {
	t.Parallel()

	for _, response := range []*ferrethttp.Response{
		{
			StatusCode: 429,
			Headers:    ferrethttp.Headers{"Content-Type": {"application/json"}},
			Body:       []byte(`{"not_error":"invalid"}`),
		},
		{
			StatusCode: 503,
			Headers:    ferrethttp.Headers{"Content-Type": {"text/plain"}},
			Body:       []byte("unavailable"),
		},
	} {
		_, err := parseTokenEndpointResponse(response, time.Now(), "", "", nil)
		if !errors.Is(err, ErrInvalidTokenResponse) {
			t.Fatalf("expected ErrInvalidTokenResponse, got %v", err)
		}
		if !strings.Contains(err.Error(), fmt.Sprint(response.StatusCode)) {
			t.Fatalf("error does not include HTTP status %d: %v", response.StatusCode, err)
		}
	}
}

func TestParseTokenOAuthErrorFormatting(t *testing.T) {
	oauthError, err := parseTokenOAuthError([]byte(`{
		"error":"invalid_grant",
		"error_description":"the grant is invalid",
		"error_uri":"https://example.com/errors/invalid-grant"
	}`), 400, nil)
	if err != nil {
		t.Fatalf("parse OAuth error: %v", err)
	}

	if got := oauthError.Error(); got !=
		"oauth2 token request: invalid_grant: the grant is invalid" {
		t.Fatalf("unexpected error text: %q", got)
	}
	if strings.Contains(oauthError.Error(), oauthError.URI) {
		t.Fatal("error string unexpectedly included error URI")
	}
}
