package object

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRequireMap(t *testing.T) {
	if _, err := RequireMap(runtime.NewObject(), "owner"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := RequireMap(runtime.None, "owner")
	if err == nil || err.Error() != "owner must be an object" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestString(t *testing.T) {
	ctx := context.Background()
	obj := runtime.NewObjectWith(map[string]runtime.Value{
		"name": runtime.NewString("ferret"),
		"bad":  runtime.NewInt(1),
	})

	got, found, err := String(ctx, obj, "name", "owner")
	if err != nil || !found || got != "ferret" {
		t.Fatalf("unexpected string lookup: %q %v %v", got, found, err)
	}

	_, found, err = String(ctx, obj, "missing", "owner")
	if err != nil || found {
		t.Fatalf("unexpected missing lookup: %v %v", found, err)
	}

	_, _, err = String(ctx, obj, "bad", "owner")
	if err == nil || err.Error() != "owner.bad must be a string" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMillisDuration(t *testing.T) {
	ctx := context.Background()
	obj := runtime.NewObjectWith(map[string]runtime.Value{
		"timeout":  runtime.NewInt(250),
		"bad":      runtime.NewString("slow"),
		"negative": runtime.NewInt(-1),
	})

	got, found, err := MillisDuration(ctx, obj, "timeout", "owner")
	if err != nil || !found || got != 250*time.Millisecond {
		t.Fatalf("unexpected duration lookup: %v %v %v", got, found, err)
	}

	_, _, err = MillisDuration(ctx, obj, "negative", "owner")
	if err == nil || !strings.Contains(err.Error(), "owner.negative must be greater than or equal to 0") {
		t.Fatalf("unexpected negative duration error: %v", err)
	}

	_, _, err = MillisDuration(ctx, obj, "bad", "owner")
	if err == nil || !strings.Contains(err.Error(), "owner.bad must be an integer number of milliseconds") {
		t.Fatalf("unexpected type error: %v", err)
	}
}

func TestAliasHelpers(t *testing.T) {
	ctx := context.Background()
	obj := runtime.NewObjectWith(map[string]runtime.Value{
		"Name":      runtime.NewString("session"),
		"http_only": runtime.True,
		"max_age":   runtime.NewInt(10),
		"bad":       runtime.NewString("bad"),
	})

	name, found, err := AliasString(ctx, obj, "name", "Name")
	if err != nil || !found || name != "session" {
		t.Fatalf("unexpected alias string: %q %v %v", name, found, err)
	}

	httpOnly, found, err := AliasBool(ctx, obj, "httpOnly", "HTTPOnly", "http_only")
	if err != nil || !found || !httpOnly {
		t.Fatalf("unexpected alias bool: %v %v %v", httpOnly, found, err)
	}

	maxAge, found, err := AliasInt(ctx, obj, "maxAge", "MaxAge", "max_age")
	if err != nil || !found || maxAge != runtime.NewInt(10) {
		t.Fatalf("unexpected alias int: %v %v %v", maxAge, found, err)
	}

	_, _, err = AliasBool(ctx, obj, "bad")
	if !errors.Is(err, runtime.ErrInvalidType) {
		t.Fatalf("expected invalid type error, got %v", err)
	}
}

func TestStringMap(t *testing.T) {
	ctx := context.Background()
	obj := runtime.NewObjectWith(map[string]runtime.Value{
		"accept": runtime.NewString("application/json"),
	})

	got, err := StringMap(ctx, obj, "headers")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["accept"] != "application/json" {
		t.Fatalf("unexpected string map: %#v", got)
	}

	bad := runtime.NewObjectWith(map[string]runtime.Value{
		"accept": runtime.NewInt(1),
	})
	_, err = StringMap(ctx, bad, "headers")
	if err == nil || err.Error() != "headers values must be strings" {
		t.Fatalf("unexpected error: %v", err)
	}
}
