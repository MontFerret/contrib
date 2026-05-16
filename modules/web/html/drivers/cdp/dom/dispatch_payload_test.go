package dom

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestValidateDispatchEventRejectsOptions(t *testing.T) {
	t.Parallel()

	err := validateDispatchEvent(runtime.DispatchEvent{
		Name:    runtime.NewString(drivers.DispatchClickEvent),
		Options: runtime.NewObject(),
	})

	if !errors.Is(err, runtime.ErrInvalidOperation) {
		t.Fatalf("expected invalid operation, got %v", err)
	}
}

func TestValidateDispatchEventRejectsUnknownEvent(t *testing.T) {
	t.Parallel()

	err := validateDispatchEvent(runtime.DispatchEvent{
		Name: runtime.NewString("Click"),
	})

	if !errors.Is(err, runtime.ErrInvalidOperation) {
		t.Fatalf("expected invalid operation, got %v", err)
	}

	expectedEvents := strings.Join(drivers.SupportedDispatchEvents(), ", ")

	if !strings.Contains(err.Error(), "supported events: "+expectedEvents) {
		t.Fatalf("expected supported event list in error, got %v", err)
	}
}

func TestParseDispatchMousePayload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	payload := runtime.NewObjectWith(map[string]runtime.Value{
		"button": runtime.NewString("right"),
		"count":  runtime.NewInt(3),
		"x":      runtime.NewFloat(10),
		"y":      runtime.NewInt(5),
	})

	params, err := parseDispatchMousePayload(ctx, drivers.DispatchClickEvent, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.Button != "right" || params.Count != 3 {
		t.Fatalf("unexpected mouse params: %#v", params)
	}

	if params.X == nil || *params.X != 10 {
		t.Fatalf("unexpected x offset: %#v", params.X)
	}

	if params.Y == nil || *params.Y != 5 {
		t.Fatalf("unexpected y offset: %#v", params.Y)
	}

	params, err = parseDispatchMousePayload(ctx, drivers.DispatchDoubleClickEvent, runtime.None)
	if err != nil {
		t.Fatalf("unexpected default error: %v", err)
	}

	if params.Button != "left" || params.Count != 2 {
		t.Fatalf("unexpected default dblclick params: %#v", params)
	}
}

func TestParseDispatchKeyboardPayload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	payload := runtime.NewObjectWith(map[string]runtime.Value{
		"keys":  runtime.NewArrayWith(runtime.NewString("Meta"), runtime.NewString("A")),
		"count": runtime.NewInt(2),
	})

	params, err := parseDispatchKeyboardPayload(ctx, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.Count != 2 || len(params.Keys) != 2 || params.Keys[0] != "Meta" || params.Keys[1] != "A" {
		t.Fatalf("unexpected keyboard params: %#v", params)
	}

	key, err := parseDispatchKeyPayload(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"key": runtime.NewString("Enter"),
	}))
	if err != nil {
		t.Fatalf("unexpected key error: %v", err)
	}

	if key != "Enter" {
		t.Fatalf("unexpected key: %q", key)
	}
}

func TestParseDispatchTypePayload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	params, err := parseDispatchTypePayload(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"text":  runtime.NewString("macbook"),
		"delay": runtime.NewInt(25),
		"clear": runtime.True,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if params.Text != "macbook" || params.Delay != 25 || !params.Clear {
		t.Fatalf("unexpected type params: %#v", params)
	}

	_, err = parseDispatchTypePayload(ctx, runtime.NewObject())
	if !errors.Is(err, runtime.ErrMissedArgument) {
		t.Fatalf("expected missed text argument, got %v", err)
	}
}

func TestParseDispatchScrollPayload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	params, err := parseDispatchScrollPayload(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"y":        runtime.NewInt(1200),
		"behavior": runtime.NewString("smooth"),
	}))
	if err != nil {
		t.Fatalf("unexpected direct scroll error: %v", err)
	}

	if params.Mode != dispatchScrollModeTo || params.Options.Top != 1200 || params.Options.Behavior != drivers.ScrollBehaviorSmooth {
		t.Fatalf("unexpected direct scroll params: %#v", params)
	}

	params, err = parseDispatchScrollPayload(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"by": runtime.NewObjectWith(map[string]runtime.Value{
			"x": runtime.NewInt(10),
			"y": runtime.NewInt(20),
		}),
	}))
	if err != nil {
		t.Fatalf("unexpected relative scroll error: %v", err)
	}

	if params.Mode != dispatchScrollModeBy || params.Options.Left != 10 || params.Options.Top != 20 {
		t.Fatalf("unexpected relative scroll params: %#v", params)
	}

	params, err = parseDispatchScrollPayload(ctx, runtime.NewObjectWith(map[string]runtime.Value{
		"intoView": runtime.True,
	}))
	if err != nil {
		t.Fatalf("unexpected into-view scroll error: %v", err)
	}

	if params.Mode != dispatchScrollModeIntoView {
		t.Fatalf("unexpected into-view params: %#v", params)
	}

	_, err = parseDispatchScrollPayload(ctx, runtime.NewObject())
	if !errors.Is(err, runtime.ErrMissedArgument) {
		t.Fatalf("expected missed coordinates, got %v", err)
	}
}
