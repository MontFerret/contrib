package lib

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes string input into normalized document", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString(" \n<book id=\"123\"><title>Hello</title><price>19</price></book>\n "))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		doc := mustRuntimeObject(t, result)
		assertObjectFieldString(t, ctx, doc, "type", "document")

		root := mustObjectFieldObject(t, ctx, doc, "root")
		assertObjectFieldString(t, ctx, root, "type", "element")
		assertObjectFieldString(t, ctx, root, "name", "book")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, root, "attrs"), "id", "123")

		children := mustObjectFieldArray(t, ctx, root, "children")
		assertArrayLen(t, ctx, children, 2)

		title := mustArrayObjectAtIndex(t, ctx, children, 0)
		assertObjectFieldString(t, ctx, title, "name", "title")
		titleChildren := mustObjectFieldArray(t, ctx, title, "children")
		assertArrayLen(t, ctx, titleChildren, 1)
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, titleChildren, 0), "value", "Hello")

		price := mustArrayObjectAtIndex(t, ctx, children, 1)
		assertObjectFieldString(t, ctx, price, "name", "price")
		priceChildren := mustObjectFieldArray(t, ctx, price, "children")
		assertArrayLen(t, ctx, priceChildren, 1)
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, priceChildren, 0), "value", "19")
	})

	t.Run("accepts binary input", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewBinary([]byte("<book><title>Hello</title></book>")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustObjectFieldObject(t, ctx, mustRuntimeObject(t, result), "root")
		assertObjectFieldString(t, ctx, root, "name", "book")
	})

	t.Run("preserves mixed content and whitespace inside elements", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("<p>Hello <b>world</b> !</p>"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustObjectFieldObject(t, ctx, mustRuntimeObject(t, result), "root")
		children := mustObjectFieldArray(t, ctx, root, "children")
		assertArrayLen(t, ctx, children, 3)

		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, children, 0), "value", "Hello ")

		bold := mustArrayObjectAtIndex(t, ctx, children, 1)
		assertObjectFieldString(t, ctx, bold, "name", "b")
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, mustObjectFieldArray(t, ctx, bold, "children"), 0), "value", "world")

		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, children, 2), "value", " !")
	})

	t.Run("preserves qualified names and skips ignored token types", func(t *testing.T) {
		input := "<?xml version=\"1.0\"?><ns:book xmlns:ns=\"urn:test\"><!--ignored--><?pi ok?><ns:title><![CDATA[Hello]]></ns:title></ns:book>"

		result, err := Decode(ctx, runtime.NewString(input))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustObjectFieldObject(t, ctx, mustRuntimeObject(t, result), "root")
		assertObjectFieldString(t, ctx, root, "name", "ns:book")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, root, "attrs"), "xmlns:ns", "urn:test")

		children := mustObjectFieldArray(t, ctx, root, "children")
		assertArrayLen(t, ctx, children, 1)

		title := mustArrayObjectAtIndex(t, ctx, children, 0)
		assertObjectFieldString(t, ctx, title, "name", "ns:title")
		titleChildren := mustObjectFieldArray(t, ctx, title, "children")
		assertArrayLen(t, ctx, titleChildren, 1)
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, titleChildren, 0), "value", "Hello")
	})

	t.Run("rejects malformed xml", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("<root><child></root>"))
		if err == nil {
			t.Fatal("expected malformed XML error")
		}

		if _, ok := err.(*core.XMLError); !ok {
			t.Fatalf("expected *core.XMLError, got %T", err)
		}
	})

	t.Run("rejects text outside root", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("hello<root/>"))
		if err == nil {
			t.Fatal("expected error for text outside root")
		}
	})

	t.Run("rejects non text input", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for non-text input")
		}
	})
}
