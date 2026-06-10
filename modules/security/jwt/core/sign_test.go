package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestSign(t *testing.T) {
	ctx := context.Background()
	key := NewHMACKey([]byte("secret"))
	claims := runtime.NewObjectWith(map[string]runtime.Value{
		"sub": runtime.NewString("user123"),
	})

	t.Run("Valid signature", func(t *testing.T) {
		opts := SignOptions{
			Algorithm: "HS256",
		}
		res, err := Sign(ctx, claims, key, opts)
		if err != nil {
			t.Fatalf("Sign() error = %v", err)
		}
		if res.String() == "" {
			t.Error("Sign() returned empty token")
		}
	})

	t.Run("Conflict subject", func(t *testing.T) {
		opts := SignOptions{
			Algorithm: "HS256",
			Subject:   "other",
		}
		_, err := Sign(ctx, claims, key, opts)
		if err == nil {
			t.Error("expected error for conflicting subject")
		}
	})

	t.Run("Missing algorithm", func(t *testing.T) {
		opts := SignOptions{}
		_, err := Sign(ctx, claims, key, opts)
		if err == nil {
			t.Error("expected error for missing algorithm")
		}
	})
}
