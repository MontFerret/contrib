package core

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestParseCompactToken(t *testing.T) {
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{"sub": "user123"}

	hBytes, _ := json.Marshal(header)
	cBytes, _ := json.Marshal(claims)

	hBase64 := base64.RawURLEncoding.EncodeToString(hBytes)
	cBase64 := base64.RawURLEncoding.EncodeToString(cBytes)
	sigBase64 := base64.RawURLEncoding.EncodeToString([]byte("signature"))

	validToken := hBase64 + "." + cBase64 + "." + sigBase64

	t.Run("Valid token", func(t *testing.T) {
		parsed, err := parseCompactToken(validToken, 1024)
		if err != nil {
			t.Fatalf("parseCompactToken() error = %v", err)
		}
		if parsed.rawHeader != hBase64 {
			t.Errorf("rawHeader mismatch")
		}
		if parsed.header["alg"] != "HS256" {
			t.Errorf("header alg mismatch")
		}
		if parsed.claims["sub"] != "user123" {
			t.Errorf("claims sub mismatch")
		}
	})

	t.Run("Too large", func(t *testing.T) {
		_, err := parseCompactToken(validToken, 5)
		if err == nil {
			t.Error("expected error for large token")
		}
	})

	t.Run("Malformed parts", func(t *testing.T) {
		_, err := parseCompactToken("part1.part2", 1024)
		if err == nil {
			t.Error("expected error for malformed token")
		}
	})

	t.Run("Empty parts", func(t *testing.T) {
		_, err := parseCompactToken("part1..part3", 1024)
		if err == nil {
			t.Error("expected error for empty parts")
		}
	})
}

func TestHeaderAlgorithm(t *testing.T) {
	tests := []struct {
		name    string
		header  map[string]any
		want    string
		wantErr bool
	}{
		{"valid", map[string]any{"alg": "HS256"}, "HS256", false},
		{"missing", map[string]any{}, "", true},
		{"invalid type", map[string]any{"alg": 123}, "", true},
		{"empty string", map[string]any{"alg": ""}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := headerAlgorithm(tt.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("headerAlgorithm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("headerAlgorithm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlgorithmAllowed(t *testing.T) {
	allowed := []string{"HS256", "RS256"}
	if !algorithmAllowed("HS256", allowed) {
		t.Error("HS256 should be allowed")
	}
	if algorithmAllowed("ES256", allowed) {
		t.Error("ES256 should not be allowed")
	}
}
