package core

import (
	"context"
	"io"
	"strings"
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

	t.Run("creates iterable iterator from reader and yields keyed events", func(t *testing.T) {
		iter, err := NewDecodeIteratorFromReader(strings.NewReader("<root id=\"1\">hi</root>"))
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

	t.Run("close marks reader iterator as done", func(t *testing.T) {
		iter, err := NewDecodeIteratorFromReader(strings.NewReader("<root/>"))
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

	t.Run("rejects nil readers", func(t *testing.T) {
		var reader *strings.Reader

		_, err := NewDecodeIteratorFromReader(reader)
		if err == nil {
			t.Fatal("expected nil reader error")
		}

		if err.Error() != "xml: reader must not be nil" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("reader constructor matches malformed xml behavior", func(t *testing.T) {
		const invalidXML = "<root></other>"

		fromString, err := NewDecodeIterator(runtime.NewString(invalidXML))
		if err != nil {
			t.Fatalf("unexpected string constructor error: %v", err)
		}

		fromReader, err := NewDecodeIteratorFromReader(strings.NewReader(invalidXML))
		if err != nil {
			t.Fatalf("unexpected reader constructor error: %v", err)
		}

		stringErr := drainIterator(ctx, fromString)
		readerErr := drainIterator(ctx, fromReader)
		if stringErr == nil || readerErr == nil {
			t.Fatalf("expected both iterators to fail, got string=%v reader=%v", stringErr, readerErr)
		}

		if stringErr.Error() != readerErr.Error() {
			t.Fatalf("expected matching errors, got string=%v reader=%v", stringErr, readerErr)
		}
	})
}

func drainIterator(ctx context.Context, iter *DecodeIterator) error {
	for {
		_, _, err := iter.Next(ctx)
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}
	}
}
