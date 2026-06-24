package core

import (
	"testing"
)

func TestSigningMethod(t *testing.T) {
	tests := []struct {
		alg     string
		code    string
		wantErr bool
	}{
		{alg: "HS256", wantErr: false, code: ""},
		{alg: "RS256", wantErr: false, code: ""},
		{alg: "ES256", wantErr: false, code: ""},
		{alg: "EdDSA", wantErr: false, code: ""},
		{alg: "none", wantErr: true, code: ErrUnexpectedAlgorithm},
		{alg: "invalid", wantErr: true, code: ErrUnsupportedAlgorithm},
	}

	for _, tt := range tests {
		t.Run(tt.alg, func(t *testing.T) {
			m, err := signingMethod(tt.alg)
			if tt.wantErr {
				if err == nil {
					t.Errorf("signingMethod() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if e, ok := err.(*Error); ok {
					if e.Code != tt.code {
						t.Errorf("signingMethod() error code = %v, want %v", e.Code, tt.code)
					}
				} else {
					t.Errorf("signingMethod() error = %v, want type *Error", err)
				}
			} else {
				if err != nil {
					t.Errorf("signingMethod() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if m.Alg() != tt.alg {
					t.Errorf("signingMethod() alg = %v, want %v", m.Alg(), tt.alg)
				}
			}
		})
	}
}

func TestNormalizeAlgorithms(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"empty", []string{}, []string{}},
		{"nil", nil, []string{}},
		{"spaces", []string{" HS256 ", ""}, []string{"HS256"}},
		{"mixed", []string{"RS256", " ", "ES256"}, []string{"RS256", "ES256"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAlgorithms(tt.in)
			if len(got) != len(tt.want) {
				t.Errorf("normalizeAlgorithms() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("normalizeAlgorithms()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestKeyMaterial(t *testing.T) {
	hmacKey := NewHMACKey([]byte("secret"))

	t.Run("HMAC sign material", func(t *testing.T) {
		got, err := signKeyMaterial(hmacKey, "HS256")
		if err != nil {
			t.Fatalf("signKeyMaterial() error = %v", err)
		}
		if string(got.([]byte)) != "secret" {
			t.Errorf("signKeyMaterial() = %v, want secret", got)
		}
	})

	t.Run("HMAC verify material", func(t *testing.T) {
		got, err := verifyKeyMaterial(hmacKey, "HS256")
		if err != nil {
			t.Fatalf("verifyKeyMaterial() error = %v", err)
		}
		if string(got.([]byte)) != "secret" {
			t.Errorf("verifyKeyMaterial() = %v, want secret", got)
		}
	})

	t.Run("HMAC wrong alg", func(t *testing.T) {
		_, err := signKeyMaterial(hmacKey, "RS256")
		if err == nil {
			t.Error("signKeyMaterial() expected error for HMAC key with RS256")
		}
	})
}
