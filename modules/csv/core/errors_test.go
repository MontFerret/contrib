package core

import (
	"errors"
	"testing"
)

func TestCSVError(t *testing.T) {
	t.Run("row only", func(t *testing.T) {
		err := newError(5, "something went wrong")
		expected := "csv: row 5: something went wrong"

		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("row and column", func(t *testing.T) {
		err := &Error{Row: 3, Column: 7, Msg: "unterminated quote"}
		expected := "csv: row 3, col 7: unterminated quote"

		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("formatted error", func(t *testing.T) {
		err := newErrorf(12, "expected %d fields but got %d", 5, 7)
		expected := "csv: row 12: expected 5 fields but got 7"

		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		inner := errors.New("inner error")
		err := &Error{Row: 1, Msg: "wrapped", Err: inner}

		if !errors.Is(err, inner) {
			t.Fatal("expected Unwrap to return inner error")
		}
	})
}
