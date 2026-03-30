package lib

import (
	"context"
	"io"
	"testing"

	"github.com/MontFerret/contrib/modules/xml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeStreamLib(t *testing.T) {
	ctx := context.Background()

	t.Run("emits normalized events with 1-based keys", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString("<root id=\"1\">hi</root>"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		first, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first event: %v", err)
		}
		assertRuntimeIntValue(t, key, 1)
		firstEvent := mustRuntimeObject(t, first)
		assertObjectFieldString(t, ctx, firstEvent, "type", "startElement")
		assertObjectFieldString(t, ctx, firstEvent, "name", "root")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, firstEvent, "attrs"), "id", "1")

		second, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading second event: %v", err)
		}
		assertRuntimeIntValue(t, key, 2)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, second), "value", "hi")

		third, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading third event: %v", err)
		}
		assertRuntimeIntValue(t, key, 3)
		thirdEvent := mustRuntimeObject(t, third)
		assertObjectFieldString(t, ctx, thirdEvent, "type", "endElement")
		assertObjectFieldString(t, ctx, thirdEvent, "name", "root")

		_, _, err = iter.Next(ctx)
		if err != io.EOF {
			t.Fatalf("expected EOF, got %v", err)
		}
	})

	t.Run("accepts binary input and skips ignored tokens", func(t *testing.T) {
		input := "<?xml version=\"1.0\"?><root><!--ignored--><?pi ok?><child/></root>"

		result, err := DecodeStream(ctx, runtime.NewBinary([]byte(input)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		event, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first event: %v", err)
		}
		assertRuntimeIntValue(t, key, 1)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, event), "name", "root")

		event, key, err = iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading second event: %v", err)
		}
		assertRuntimeIntValue(t, key, 2)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, event), "name", "child")

		event, key, err = iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading third event: %v", err)
		}
		assertRuntimeIntValue(t, key, 3)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, event), "name", "child")

		event, key, err = iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading fourth event: %v", err)
		}
		assertRuntimeIntValue(t, key, 4)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, event), "name", "root")
	})

	t.Run("preserves qualified names and namespace attrs", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString("<ns:book xmlns:ns=\"urn:test\"><ns:title/></ns:book>"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		first, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading first event: %v", err)
		}

		firstEvent := mustRuntimeObject(t, first)
		assertObjectFieldString(t, ctx, firstEvent, "name", "ns:book")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, firstEvent, "attrs"), "xmlns:ns", "urn:test")

		second, _, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected error reading second event: %v", err)
		}
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, second), "name", "ns:title")
	})

	t.Run("surfaces malformed xml on the failing iteration", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString("<root><child></root>"))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		if _, _, err := iter.Next(ctx); err != nil {
			t.Fatalf("unexpected error on first event: %v", err)
		}

		if _, _, err := iter.Next(ctx); err != nil {
			t.Fatalf("unexpected error on second event: %v", err)
		}

		_, _, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("expected iteration error")
		}

		if _, ok := err.(*core.Error); !ok {
			t.Fatalf("expected *core.XMLError, got %T", err)
		}
	})

	t.Run("rejects text outside root during iteration", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString("hello<root/>"))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		_, _, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("expected text outside root error")
		}

		if _, ok := err.(*core.Error); !ok {
			t.Fatalf("expected *core.XMLError, got %T", err)
		}
	})

	t.Run("rejects empty documents during iteration", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString(" \n<!--ignored-->\n "))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iter := mustIterate(t, ctx, result)
		_, _, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("expected empty document error")
		}

		if _, ok := err.(*core.Error); !ok {
			t.Fatalf("expected *core.XMLError, got %T", err)
		}
	})

	t.Run("rejects multiple root elements during iteration", func(t *testing.T) {
		result, err := DecodeStream(ctx, runtime.NewString("<first/><second/>"))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iter := mustIterate(t, ctx, result)

		if _, _, err := iter.Next(ctx); err != nil {
			t.Fatalf("unexpected error on first event: %v", err)
		}

		if _, _, err := iter.Next(ctx); err != nil {
			t.Fatalf("unexpected error on second event: %v", err)
		}

		_, _, err = iter.Next(ctx)
		if err == nil {
			t.Fatal("expected multiple root error")
		}

		if _, ok := err.(*core.Error); !ok {
			t.Fatalf("expected *core.XMLError, got %T", err)
		}
	})

	t.Run("rejects non text input", func(t *testing.T) {
		_, err := DecodeStream(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for non-text input")
		}
	})
}
