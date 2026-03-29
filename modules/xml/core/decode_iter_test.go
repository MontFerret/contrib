package core

import (
	"context"
	"io"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeIteratorCore(t *testing.T) {
	ctx := context.Background()

	t.Run("creates iterable iterator and yields keyed events", func(t *testing.T) {
		iter, err := NewDecodeIterator(runtime.NewString("<root id=\"1\">hi</root>"))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		iterated, err := iter.Iterate(ctx)
		if err != nil {
			t.Fatalf("unexpected iterate error: %v", err)
		}
		if iterated != iter {
			t.Fatal("expected iterator to return itself from Iterate")
		}

		first, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected first Next error: %v", err)
		}
		assertRuntimeIntValue(t, key, 1)
		firstEvent := mustRuntimeObject(t, first)
		assertObjectFieldString(t, ctx, firstEvent, "type", eventTypeStartElement)
		assertObjectFieldString(t, ctx, firstEvent, "name", "root")
		assertObjectFieldString(t, ctx, mustObjectFieldObject(t, ctx, firstEvent, "attrs"), "id", "1")

		second, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected second Next error: %v", err)
		}
		assertRuntimeIntValue(t, key, 2)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, second), "value", "hi")

		third, key, err := iter.Next(ctx)
		if err != nil {
			t.Fatalf("unexpected third Next error: %v", err)
		}
		assertRuntimeIntValue(t, key, 3)
		assertObjectFieldString(t, ctx, mustRuntimeObject(t, third), "type", eventTypeEndElement)

		_, _, err = iter.Next(ctx)
		if err != io.EOF {
			t.Fatalf("expected EOF, got %v", err)
		}
	})

	t.Run("close marks iterator as done", func(t *testing.T) {
		iter, err := NewDecodeIterator(runtime.NewString("<root/>"))
		if err != nil {
			t.Fatalf("unexpected constructor error: %v", err)
		}

		if err := iter.Close(); err != nil {
			t.Fatalf("unexpected close error: %v", err)
		}

		_, _, err = iter.Next(ctx)
		if err != io.EOF {
			t.Fatalf("expected EOF after close, got %v", err)
		}
	})
}
