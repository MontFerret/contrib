package core

import (
	"io"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeCursor(t *testing.T) {
	t.Run("iterates normalized events and skips ignored tokens", func(t *testing.T) {
		cursor := newDecodeCursor(runtime.NewString("<?xml version=\"1.0\"?><ns:book xmlns:ns=\"urn:test\"><!--ignored--><![CDATA[hi]]><child/></ns:book>"))

		first, err := cursor.Next()
		if err != nil {
			t.Fatalf("unexpected first event error: %v", err)
		}
		if first.kind != decodeEventStart || first.name != "ns:book" {
			t.Fatalf("unexpected first event: %+v", first)
		}
		assertObjectFieldString(t, t.Context(), first.attrs, "xmlns:ns", "urn:test")

		second, err := cursor.Next()
		if err != nil {
			t.Fatalf("unexpected second event error: %v", err)
		}
		if second.kind != decodeEventText || second.text != "hi" {
			t.Fatalf("unexpected second event: %+v", second)
		}

		third, err := cursor.Next()
		if err != nil {
			t.Fatalf("unexpected third event error: %v", err)
		}
		if third.kind != decodeEventStart || third.name != "child" {
			t.Fatalf("unexpected third event: %+v", third)
		}

		fourth, err := cursor.Next()
		if err != nil {
			t.Fatalf("unexpected fourth event error: %v", err)
		}
		if fourth.kind != decodeEventEnd || fourth.name != "child" {
			t.Fatalf("unexpected fourth event: %+v", fourth)
		}

		fifth, err := cursor.Next()
		if err != nil {
			t.Fatalf("unexpected fifth event error: %v", err)
		}
		if fifth.kind != decodeEventEnd || fifth.name != "ns:book" {
			t.Fatalf("unexpected fifth event: %+v", fifth)
		}

		_, err = cursor.Next()
		if err != io.EOF {
			t.Fatalf("expected EOF, got %v", err)
		}
	})

	t.Run("rejects text outside root", func(t *testing.T) {
		cursor := newDecodeCursor(runtime.NewString("hello<root/>"))
		_, err := cursor.Next()
		if err == nil {
			t.Fatal("expected error for text outside root")
		}
	})

	t.Run("rejects empty documents", func(t *testing.T) {
		cursor := newDecodeCursor(runtime.NewString(" \n<!--ignored-->\n "))
		_, err := cursor.Next()
		if err == nil {
			t.Fatal("expected empty document error")
		}
	})

	t.Run("rejects multiple root elements", func(t *testing.T) {
		cursor := newDecodeCursor(runtime.NewString("<first/><second/>"))

		if _, err := cursor.Next(); err != nil {
			t.Fatalf("unexpected first event error: %v", err)
		}
		if _, err := cursor.Next(); err != nil {
			t.Fatalf("unexpected second event error: %v", err)
		}

		_, err := cursor.Next()
		if err == nil {
			t.Fatal("expected multiple root error")
		}
	})

	t.Run("rejects mismatched closing tags", func(t *testing.T) {
		cursor := newDecodeCursor(runtime.NewString("<root></other>"))

		if _, err := cursor.Next(); err != nil {
			t.Fatalf("unexpected start event error: %v", err)
		}

		_, err := cursor.Next()
		if err == nil {
			t.Fatal("expected mismatched closing tag error")
		}
	})
}
