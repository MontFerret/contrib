package core

import (
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestResolveToken(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		val := runtime.NewString("abc")
		res, err := ResolveToken(val)
		if err != nil {
			t.Fatalf("ResolveToken() error = %v", err)
		}
		if res.String() != "abc" {
			t.Errorf("got %v, want abc", res.String())
		}
	})

	t.Run("Binary", func(t *testing.T) {
		val := runtime.NewBinary([]byte("abc"))
		res, err := ResolveToken(val)
		if err != nil {
			t.Fatalf("ResolveToken() error = %v", err)
		}
		if res.String() != "abc" {
			t.Errorf("got %v, want abc", res.String())
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := ResolveToken(runtime.True)
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})
}

func TestResolveSecret(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		val := runtime.NewString("abc")
		res, err := ResolveSecret(val)
		if err != nil {
			t.Fatalf("ResolveSecret() error = %v", err)
		}
		if string(res) != "abc" {
			t.Errorf("got %v, want abc", string(res))
		}
	})

	t.Run("Binary", func(t *testing.T) {
		val := runtime.NewBinary([]byte("abc"))
		res, err := ResolveSecret(val)
		if err != nil {
			t.Fatalf("ResolveSecret() error = %v", err)
		}
		if string(res) != "abc" {
			t.Errorf("got %v, want abc", string(res))
		}
	})
}
