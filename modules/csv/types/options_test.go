package types

import (
	"encoding/csv"
	"strings"
	"testing"
)

func TestValidateSingleRuneOption(t *testing.T) {
	t.Run("accepts valid delimiter and comment runes", func(t *testing.T) {
		delimiter, err := validateSingleRuneOption("delimiter", ";")
		if err != nil {
			t.Fatalf("unexpected delimiter error: %v", err)
		}

		if delimiter != ';' {
			t.Fatalf("expected ';', got %q", delimiter)
		}

		comment, err := validateSingleRuneOption("comment", "#")
		if err != nil {
			t.Fatalf("unexpected comment error: %v", err)
		}

		if comment != '#' {
			t.Fatalf("expected '#', got %q", comment)
		}
	})

	t.Run("rejects invalid delimiter characters", func(t *testing.T) {
		cases := []string{"\n", "\r", "\x00", "\"", "\uFFFD", string([]byte{0xff})}

		for _, value := range cases {
			_, err := validateSingleRuneOption("delimiter", value)
			if err == nil {
				t.Fatalf("expected error for delimiter %q", value)
			}

			if !strings.Contains(err.Error(), "invalid delimiter character") {
				t.Fatalf("expected invalid delimiter error, got %v", err)
			}
		}
	})

	t.Run("rejects invalid comment characters", func(t *testing.T) {
		cases := []string{"\n", "\r", "\x00", "\"", "\uFFFD"}

		for _, value := range cases {
			_, err := validateSingleRuneOption("comment", value)
			if err == nil {
				t.Fatalf("expected error for comment %q", value)
			}

			if !strings.Contains(err.Error(), "invalid comment character") {
				t.Fatalf("expected invalid comment error, got %v", err)
			}
		}
	})

	t.Run("rejects multi-rune values", func(t *testing.T) {
		_, err := validateSingleRuneOption("delimiter", "||")
		if err == nil {
			t.Fatal("expected error for multi-rune delimiter")
		}

		if !strings.Contains(err.Error(), "exactly one Unicode character") {
			t.Fatalf("expected multi-rune validation error, got %v", err)
		}
	})
}

func TestOptionsApplyToReader(t *testing.T) {
	t.Run("rejects comment equal to delimiter", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Delimiter = ";"
		opts.Comment = ";"

		reader := csv.NewReader(strings.NewReader("a;b"))
		err := opts.ApplyToReader(reader)
		if err == nil {
			t.Fatal("expected error for equal delimiter and comment")
		}

		if !strings.Contains(err.Error(), "comment must differ from delimiter") {
			t.Fatalf("expected equal delimiter/comment error, got %v", err)
		}
	})

	t.Run("rejects comment equal to default delimiter", func(t *testing.T) {
		opts := DefaultOptions()
		opts.Comment = ","

		reader := csv.NewReader(strings.NewReader("a,b"))
		err := opts.ApplyToReader(reader)
		if err == nil {
			t.Fatal("expected error for comment equal to default delimiter")
		}

		if !strings.Contains(err.Error(), "comment must differ from delimiter") {
			t.Fatalf("expected equal delimiter/comment error, got %v", err)
		}
	})
}
