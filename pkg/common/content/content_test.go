package content

import (
	"errors"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestStringOrBinary(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got, err := StringOrBinary(runtime.NewString("value"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "value" {
			t.Fatalf("unexpected value: %q", got.String())
		}
	})

	t.Run("binary", func(t *testing.T) {
		got, err := StringOrBinary(runtime.NewBinary([]byte("value")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "value" {
			t.Fatalf("unexpected value: %q", got.String())
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		_, err := StringOrBinary(runtime.NewInt(1))
		if !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected invalid type error, got %v", err)
		}
	})
}

func TestBytesFromStringOrBinary(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got, err := BytesFromStringOrBinary(runtime.NewString("value"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != "value" {
			t.Fatalf("unexpected value: %q", string(got))
		}
	})

	t.Run("binary", func(t *testing.T) {
		got, err := BytesFromStringOrBinary(runtime.NewBinary([]byte{0, 1, 2}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 3 || got[0] != 0 || got[1] != 1 || got[2] != 2 {
			t.Fatalf("unexpected bytes: %v", got)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		_, err := BytesFromStringOrBinary(runtime.NewInt(1))
		if !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected invalid type error, got %v", err)
		}
	})
}
