package core

import (
	"strings"
	"testing"
)

func TestAuthorizationHeaderCanonicalizesBearer(t *testing.T) {
	header, err := AuthorizationHeader(&TokenSet{
		AccessToken: "abc-._~+/==",
		TokenType:   "bEaReR",
	})
	if err != nil {
		t.Fatalf("authorization header: %v", err)
	}

	if got := header["Authorization"]; got != "Bearer abc-._~+/==" {
		t.Fatalf("unexpected Authorization header: %q", got)
	}
}

func TestAuthorizationHeaderRejectsUnsafeOrNonBearerTokens(t *testing.T) {
	tests := []TokenSet{
		{},
		{AccessToken: "access", TokenType: "MAC"},
		{AccessToken: "access", TokenType: " Bearer"},
		{AccessToken: "access\r\nInjected: value", TokenType: "Bearer"},
		{AccessToken: "abc=def", TokenType: "Bearer"},
	}

	for _, token := range tests {
		_, err := AuthorizationHeader(&token)
		if err == nil {
			t.Fatalf("expected rejection for %#v", token)
		}
		for _, secret := range []string{token.AccessToken, "\r", "\n"} {
			if secret != "" && strings.Contains(err.Error(), secret) {
				t.Fatalf("error leaked unsafe token data: %v", err)
			}
		}
	}
}
