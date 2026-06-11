package core

import (
	"context"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestVerify(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	secret := []byte("secret")
	key := NewHMACKey(secret)

	// Create a token
	claims := runtime.NewObjectWith(map[string]runtime.Value{
		"sub": runtime.NewString("user123"),
	})
	signOpts := SignOptions{
		Algorithm: "HS256",
	}
	tokenVal, _ := Sign(ctx, claims, key, signOpts)
	token := tokenVal.(runtime.String)

	t.Run("Valid token", func(t *testing.T) {
		verifyOpts := VerifyOptions{
			Algorithms: []string{"HS256"},
		}
		res, err := Verify(ctx, cfg, token, key, verifyOpts)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}
		obj := res.(*runtime.Object)
		verified, _ := obj.Get(ctx, runtime.NewString("verified"))
		if verified != runtime.True {
			t.Error("expected verified true")
		}
	})

	t.Run("Invalid signature", func(t *testing.T) {
		wrongKey := NewHMACKey([]byte("wrong"))
		verifyOpts := VerifyOptions{
			Algorithms: []string{"HS256"},
		}
		_, err := Verify(ctx, cfg, token, wrongKey, verifyOpts)
		if err == nil {
			t.Error("expected error for wrong key")
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		signOptsExp := SignOptions{
			Algorithm: "HS256",
			ExpiresIn: 10,
		}
		expiredTokenVal, _ := Sign(ctx, claims, key, signOptsExp)
		expiredToken := expiredTokenVal.(runtime.String)

		verifyOpts := VerifyOptions{
			Algorithms: []string{"HS256"},
			Now:        time.Now().Unix() + 100, // simulate future
		}
		_, err := Verify(ctx, cfg, expiredToken, key, verifyOpts)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("Algorithm not allowed", func(t *testing.T) {
		verifyOpts := VerifyOptions{
			Algorithms: []string{"RS256"},
		}
		_, err := Verify(ctx, cfg, token, key, verifyOpts)
		if err == nil {
			t.Error("expected error for disallowed algorithm")
		}
	})
}

func TestInspect(t *testing.T) {
	ctx := context.Background()
	cfg := Config{}
	token := runtime.NewString("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")

	res, err := Inspect(ctx, cfg, token)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	obj := res.(*runtime.Object)
	verified, _ := obj.Get(ctx, runtime.NewString("verified"))
	if verified != runtime.False {
		t.Error("inspect should return verified=false")
	}
}

func TestValidateTimeClaims(t *testing.T) {
	now := time.Now().Unix()

	t.Run("Valid exp", func(t *testing.T) {
		claims := map[string]any{"exp": float64(now + 100)}
		err := validateTimeClaims(claims, now, 0, 0)
		if err != nil {
			t.Errorf("validateTimeClaims() error = %v", err)
		}
	})

	t.Run("Expired", func(t *testing.T) {
		claims := map[string]any{"exp": float64(now - 100)}
		err := validateTimeClaims(claims, now, 0, 0)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("Within leeway", func(t *testing.T) {
		claims := map[string]any{"exp": float64(now - 30)}
		err := validateTimeClaims(claims, now, 60, 0)
		if err != nil {
			t.Errorf("validateTimeClaims() error = %v, expected leeway to cover it", err)
		}
	})
}
