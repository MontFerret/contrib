package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDecodeLib(t *testing.T) {
	ctx := context.Background()

	t.Run("too few args", func(t *testing.T) {
		_, err := Decode(ctx)
		if err == nil {
			t.Fatal("expected error for no args")
		}
	})

	t.Run("too many args", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("a"), runtime.NewObject(), runtime.NewString("extra"))
		if err == nil {
			t.Fatal("expected error for too many args")
		}
	})

	t.Run("wrong arg type", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewInt(42))
		if err == nil {
			t.Fatal("expected error for wrong arg type")
		}
	})

	t.Run("valid call string only", func(t *testing.T) {
		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertRuntimeArrayLen(t, ctx, arr, 1)
	})

	t.Run("valid call with options", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewString(";"),
		})

		result, err := Decode(ctx, runtime.NewString("name;age\nAlice;30"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertRuntimeArrayLen(t, ctx, arr, 1)
	})

	t.Run("wrong options type", func(t *testing.T) {
		_, err := Decode(ctx, runtime.NewString("a,b"), runtime.NewString("not-a-map"))
		if err == nil {
			t.Fatal("expected error for wrong options type")
		}
	})

	t.Run("rejects unknown options", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"unknown": runtime.True,
		})

		_, err := Decode(ctx, runtime.NewString("name\nAlice"), opts)
		if err == nil {
			t.Fatal("expected unknown option error")
		}
	})

	t.Run("propagates canceled option decoding context", func(t *testing.T) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := Decode(canceled, runtime.NewString("name\nAlice"), runtime.NewObject())
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation, got %v", err)
		}
	})

	t.Run("camelCase options are decoded", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"skipEmpty":  runtime.True,
			"inferTypes": runtime.True,
			"nullValues": runtime.NewArrayWith(runtime.NewString("null")),
		})

		result, err := Decode(ctx, runtime.NewString("name,age\nAlice,30\n,\nBob,null"), opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		arr := mustRuntimeArray(t, result)
		assertRuntimeArrayLen(t, ctx, arr, 2)

		first := mustObjectAtIndex(t, ctx, arr, 0)
		age, err := first.Get(ctx, runtime.NewString("age"))
		if err != nil {
			t.Fatalf("unexpected error getting age: %v", err)
		}
		assertRuntimeIntValue(t, age, 30)

		second := mustObjectAtIndex(t, ctx, arr, 1)
		age, err = second.Get(ctx, runtime.NewString("age"))
		if err != nil {
			t.Fatalf("unexpected error getting null age: %v", err)
		}
		if age != runtime.None {
			t.Fatalf("expected null age to decode as None, got %v", age)
		}
	})

	t.Run("invalid delimiter rune returns error", func(t *testing.T) {
		opts := runtime.NewObjectWith(map[string]runtime.Value{
			"delimiter": runtime.NewString("\n"),
		})

		_, err := Decode(ctx, runtime.NewString("name,age\nAlice,30"), opts)
		if err == nil {
			t.Fatal("expected error for invalid delimiter")
		}
	})
}
