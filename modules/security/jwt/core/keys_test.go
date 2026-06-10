package core

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func generateRSAKeys(t *testing.T) (string, string) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return string(privPEM), string(pubPEM)
}

func generateECDSAKeys(t *testing.T) (string, string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return string(privPEM), string(pubPEM)
}

func generateEd25519Keys(t *testing.T) (string, string) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	return string(privPEM), string(pubPEM)
}

func TestHMACKey(t *testing.T) {
	secret := []byte("my-secret")
	key := NewHMACKey(secret)
	if string(key.Secret()) != string(secret) {
		t.Errorf("Secret() = %v, want %v", key.Secret(), secret)
	}
	if key.ResourceID() == 0 {
		t.Error("ResourceID() should not be 0")
	}
	if key.String() != "<security.jwt.hmac_key>" {
		t.Errorf("String() = %v, want <security.jwt.hmac_key>", key.String())
	}
}

func TestRSAKeys(t *testing.T) {
	privPEM, pubPEM := generateRSAKeys(t)

	t.Run("Private Key", func(t *testing.T) {
		priv, err := NewPrivateKey(privPEM)
		if err != nil {
			t.Fatalf("NewPrivateKey() error = %v", err)
		}
		if priv.family != familyRSA {
			t.Errorf("family = %v, want RSA", priv.family)
		}
		if priv.rsaKey == nil {
			t.Error("rsaKey is nil")
		}
	})

	t.Run("Public Key", func(t *testing.T) {
		pub, err := NewPublicKey(pubPEM)
		if err != nil {
			t.Fatalf("NewPublicKey() error = %v", err)
		}
		if pub.family != familyRSA {
			t.Errorf("family = %v, want RSA", pub.family)
		}
		if pub.rsaKey == nil {
			t.Error("rsaKey is nil")
		}
	})
}

func TestECDSAKeys(t *testing.T) {
	privPEM, pubPEM := generateECDSAKeys(t)

	t.Run("Private Key", func(t *testing.T) {
		priv, err := NewPrivateKey(privPEM)
		if err != nil {
			t.Fatalf("NewPrivateKey() error = %v", err)
		}
		if priv.family != familyECDSA {
			t.Errorf("family = %v, want ECDSA", priv.family)
		}
		if priv.ecKey == nil {
			t.Error("ecKey is nil")
		}
	})

	t.Run("Public Key", func(t *testing.T) {
		pub, err := NewPublicKey(pubPEM)
		if err != nil {
			t.Fatalf("NewPublicKey() error = %v", err)
		}
		if pub.family != familyECDSA {
			t.Errorf("family = %v, want ECDSA", pub.family)
		}
		if pub.ecKey == nil {
			t.Error("ecKey is nil")
		}
	})
}

func TestEd25519Keys(t *testing.T) {
	privPEM, pubPEM := generateEd25519Keys(t)

	t.Run("Private Key", func(t *testing.T) {
		priv, err := NewPrivateKey(privPEM)
		if err != nil {
			t.Fatalf("NewPrivateKey() error = %v", err)
		}
		if priv.family != familyEd25519 {
			t.Errorf("family = %v, want Ed25519", priv.family)
		}
		if priv.edKey == nil {
			t.Error("edKey is nil")
		}
	})

	t.Run("Public Key", func(t *testing.T) {
		pub, err := NewPublicKey(pubPEM)
		if err != nil {
			t.Fatalf("NewPublicKey() error = %v", err)
		}
		if pub.family != familyEd25519 {
			t.Errorf("family = %v, want Ed25519", pub.family)
		}
		if pub.edKey == nil {
			t.Error("edKey is nil")
		}
	})
}

func TestVerifyKeyMatchesAlgorithm(t *testing.T) {
	hmacKey := NewHMACKey([]byte("secret"))
	rsaPrivPEM, rsaPubPEM := generateRSAKeys(t)
	rsaPriv, _ := NewPrivateKey(rsaPrivPEM)
	rsaPub, _ := NewPublicKey(rsaPubPEM)

	tests := []struct {
		name    string
		key     any
		alg     string
		wantErr bool
	}{
		{"HMAC-HS256", hmacKey, "HS256", false},
		{"HMAC-RS256", hmacKey, "RS256", true},
		{"RSA-RS256-Priv", rsaPriv, "RS256", false},
		{"RSA-RS256-Pub", rsaPub, "RS256", false},
		{"RSA-HS256", rsaPriv, "HS256", true},
		{"RSA-ES256", rsaPriv, "ES256", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyKeyMatchesAlgorithm(tt.key, tt.alg)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyKeyMatchesAlgorithm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
