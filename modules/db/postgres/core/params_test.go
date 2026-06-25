package core

import (
	"context"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestParseParams(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2026, 6, 25, 10, 30, 0, 0, time.UTC)

	params, err := parseParams(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"params": runtime.NewArrayWith(
			runtime.NewInt(42),
			runtime.NewFloat(4.25),
			runtime.NewString("ferret"),
			runtime.True,
			runtime.NewBinary([]byte("bin")),
			runtime.NewDateTime(now),
			runtime.None,
		),
	}))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(params) != 7 {
		t.Fatalf("expected 7 params, got %d", len(params))
	}
	if params[0] != int64(42) {
		t.Fatalf("unexpected int param: %v", params[0])
	}
	if params[1] != float64(4.25) {
		t.Fatalf("unexpected float param: %v", params[1])
	}
	if params[2] != "ferret" {
		t.Fatalf("unexpected string param: %v", params[2])
	}
	if params[3] != true {
		t.Fatalf("unexpected bool param: %v", params[3])
	}
	if string(params[4].([]byte)) != "bin" {
		t.Fatalf("unexpected binary param: %v", params[4])
	}
	if !params[5].(time.Time).Equal(now) {
		t.Fatalf("unexpected datetime param: %v", params[5])
	}
	if params[6] != nil {
		t.Fatalf("unexpected none param: %v", params[6])
	}
}

func TestParseParamsErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, err := parseParams(ctx, runtime.NewString("invalid"))
	assertErrorContains(t, err, "query params must be an object")

	_, err = parseParams(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"params": runtime.NewString("invalid"),
	}))
	assertErrorContains(t, err, "params must be an array")

	_, err = parseParams(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"params": runtime.NewArrayWith(runtime.NewObject()),
	}))
	assertErrorContains(t, err, "unsupported param type Object")
}
