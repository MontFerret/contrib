package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestHelperLib(t *testing.T) {
	ctx := context.Background()

	rootElement := xmlElement("book", map[string]runtime.Value{
		"id":    runtime.NewString("123"),
		"genre": runtime.NewString("fiction"),
	}, xmlText("Hello "), xmlElement("title", nil, xmlText("world")), xmlText("!"))

	document := xmlDocument(rootElement)
	textNode := xmlText("standalone")

	t.Run("root unwraps documents and hides text nodes", func(t *testing.T) {
		result, err := Root(ctx, document)
		if err != nil {
			t.Fatalf("unexpected document root error: %v", err)
		}
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, result), "name", "book")

		result, err = Root(ctx, rootElement)
		if err != nil {
			t.Fatalf("unexpected element root error: %v", err)
		}
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, result), "name", "book")

		result, err = Root(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected text root error: %v", err)
		}
		if result != runtime.None {
			t.Fatalf("expected runtime.None, got %T", result)
		}
	})

	t.Run("text concatenates descendant text in document order", func(t *testing.T) {
		result, err := Text(ctx, document)
		if err != nil {
			t.Fatalf("unexpected document text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "Hello world!")

		result, err = Text(ctx, rootElement)
		if err != nil {
			t.Fatalf("unexpected element text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "Hello world!")

		result, err = Text(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected text-node text error: %v", err)
		}
		assertRuntimeStringValue(t, result, "standalone")
	})

	t.Run("attr delegates through document roots and returns none when missing", func(t *testing.T) {
		result, err := Attr(ctx, document, runtime.NewString("id"))
		if err != nil {
			t.Fatalf("unexpected document attr error: %v", err)
		}
		assertRuntimeStringValue(t, result, "123")

		result, err = Attr(ctx, rootElement, runtime.NewString("genre"))
		if err != nil {
			t.Fatalf("unexpected element attr error: %v", err)
		}
		assertRuntimeStringValue(t, result, "fiction")

		result, err = Attr(ctx, rootElement, runtime.NewString("missing"))
		if err != nil {
			t.Fatalf("unexpected missing attr error: %v", err)
		}
		if result != runtime.None {
			t.Fatalf("expected runtime.None for missing attr, got %T", result)
		}

		result, err = Attr(ctx, textNode, runtime.NewString("missing"))
		if err != nil {
			t.Fatalf("unexpected text attr error: %v", err)
		}
		if result != runtime.None {
			t.Fatalf("expected runtime.None for text attr, got %T", result)
		}
	})

	t.Run("children delegate through document roots and return empty arrays for text", func(t *testing.T) {
		result, err := Children(ctx, document)
		if err != nil {
			t.Fatalf("unexpected document children error: %v", err)
		}
		children := mustRuntimeArray(t, result)
		assertArrayLen(t, ctx, children, 3)
		assertObjectFieldString(t, ctx, mustArrayObjectAtIndex(t, ctx, children, 1), "name", "title")

		result, err = Children(ctx, rootElement)
		if err != nil {
			t.Fatalf("unexpected element children error: %v", err)
		}
		assertArrayLen(t, ctx, mustRuntimeArray(t, result), 3)

		result, err = Children(ctx, textNode)
		if err != nil {
			t.Fatalf("unexpected text children error: %v", err)
		}
		assertArrayLen(t, ctx, mustRuntimeArray(t, result), 0)
	})

	t.Run("rejects non xml values and unsupported node types", func(t *testing.T) {
		if _, err := Root(ctx, runtime.NewInt(42)); err == nil || !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected runtime invalid type error, got %v", err)
		}

		if _, err := Text(ctx, runtime.NewInt(42)); err == nil || !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected runtime invalid type error, got %v", err)
		}

		if _, err := Children(ctx, runtime.NewInt(42)); err == nil || !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected runtime invalid type error, got %v", err)
		}

		if _, err := Attr(ctx, runtime.NewInt(42), runtime.NewString("id")); err == nil || !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected runtime invalid type error, got %v", err)
		}

		invalidNode := runtime.NewObjectWith(map[string]runtime.Value{
			"type": runtime.NewString("startElement"),
		})
		if _, err := Root(ctx, invalidNode); err == nil {
			t.Fatal("expected invalid node type error")
		} else {
			var xmlErr *core.XMLError
			if !errors.As(err, &xmlErr) {
				t.Fatalf("expected *core.XMLError, got %T", err)
			}
		}
	})

	t.Run("attr requires a string attribute name", func(t *testing.T) {
		if _, err := Attr(ctx, rootElement, runtime.NewInt(1)); err == nil || !errors.Is(err, runtime.ErrInvalidType) {
			t.Fatalf("expected runtime invalid type error, got %v", err)
		}
	})
}
