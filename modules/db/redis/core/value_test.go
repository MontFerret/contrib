package core

import (
	"context"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRedisValueToRuntimeScalarsAndNestedValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	value, err := redisValueToRuntime(ctx, []any{
		nil,
		"text",
		int64(42),
		1.5,
		true,
		[]any{"nested", int64(2)},
		map[any]any{"name": "Tim", "active": true},
	})
	if err != nil {
		t.Fatalf("unexpected conversion error: %v", err)
	}

	list, ok := value.(runtime.List)
	if !ok {
		t.Fatalf("expected list, got %T", value)
	}
	length, err := list.Length(ctx)
	if err != nil || length != 7 {
		t.Fatalf("unexpected list length: %v, %v", length, err)
	}

	first, _ := list.At(ctx, runtime.ZeroInt)
	if first != runtime.None {
		t.Fatalf("expected nil to become NONE, got %v", first)
	}

	nestedMapValue, _ := list.At(ctx, runtime.NewInt(6))
	nestedMap, ok := nestedMapValue.(runtime.Map)
	if !ok {
		t.Fatalf("expected object, got %T", nestedMapValue)
	}
	if objectValue(t, ctx, nestedMap, "name") != runtime.NewString("Tim") {
		t.Fatalf("unexpected map conversion: %v", nestedMap)
	}
}

func TestRedisValueToRuntimeBigIntegerRange(t *testing.T) {
	t.Parallel()

	value, err := redisValueToRuntime(context.Background(), big.NewInt(math.MaxInt64))
	if err != nil || value != runtime.NewInt64(math.MaxInt64) {
		t.Fatalf("unexpected in-range big integer: %v, %v", value, err)
	}

	tooLarge := new(big.Int).Add(big.NewInt(math.MaxInt64), big.NewInt(1))
	_, err = redisValueToRuntime(context.Background(), tooLarge)
	assertErrorContains(t, err, "exceeds Ferret integer range")
}

func TestRedisValueToRuntimeRejectsMapKeysAndNestedErrors(t *testing.T) {
	t.Parallel()

	_, err := redisValueToRuntime(context.Background(), map[any]any{int64(1): "value"})
	assertErrorContains(t, err, "unsupported Redis map key type")

	wantErr := errors.New("ERR nested")
	_, err = redisValueToRuntime(context.Background(), []any{"ok", wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected nested Redis error, got %v", err)
	}
}

func TestRedisValueToRuntimeRejectsUnsupportedNativeType(t *testing.T) {
	t.Parallel()

	_, err := redisValueToRuntime(context.Background(), struct{}{})
	assertErrorContains(t, err, "unsupported Redis response type")
}
