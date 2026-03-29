package core

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeCore(t *testing.T) {
	ctx := context.Background()

	t.Run("decodes nested document into normalized structure", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("<book id=\"123\"><title>Hello</title><price>19</price></book>"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		doc := mustRuntimeObject(t, result)
		assertObjectFieldString(t, ctx, doc, "type", nodeTypeDocument)

		root := mustObjectFieldObject(t, ctx, doc, "root")
		assertObjectFieldString(t, ctx, root, "name", "book")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, root, "attrs"), "id", "123")

		children := mustObjectFieldArray(t, ctx, root, "children")
		assertArrayLen(t, ctx, children, 2)

		title := mustArrayObjectAtIndex(t, ctx, children, 0)
		assertObjectFieldString(t, ctx, title, "name", "title")
		titleChildren := mustObjectFieldArray(t, ctx, title, "children")
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, titleChildren, 0), "value", "Hello")
	})

	t.Run("preserves mixed content and qualified names", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("<ns:p xmlns:ns=\"urn:test\">Hello <b>world</b> !</ns:p>"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		root := mustObjectFieldObject(t, ctx, mustRuntimeObject(t, result), "root")
		assertObjectFieldString(t, ctx, root, "name", "ns:p")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, root, "attrs"), "xmlns:ns", "urn:test")

		children := mustObjectFieldArray(t, ctx, root, "children")
		assertArrayLen(t, ctx, children, 3)
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, children, 0), "value", "Hello ")
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, children, 2), "value", " !")
	})

	t.Run("rejects empty documents", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString(" \n<!--ignored-->\n "))
		if err == nil {
			t.Fatal("expected empty document error")
		}
	})

	t.Run("rejects multiple root elements", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("<first/><second/>"))
		if err == nil {
			t.Fatal("expected multiple roots error")
		}
	})
}
