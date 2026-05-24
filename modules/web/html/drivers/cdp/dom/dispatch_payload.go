package dom

import (
	"context"
	"strings"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type (
	dispatchPayload struct {
		fields runtime.Map
	}

	dispatchKeyboardParams struct {
		Keys  []runtime.String
		Count runtime.Int
	}

	dispatchTypeParams struct {
		Text  string
		Clear bool
		Delay runtime.Int
	}

	dispatchScrollMode string

	dispatchScrollParams struct {
		Mode    dispatchScrollMode
		Options drivers.ScrollOptions
	}
)

const (
	dispatchScrollModeTo       dispatchScrollMode = "to"
	dispatchScrollModeBy       dispatchScrollMode = "by"
	dispatchScrollModeIntoView dispatchScrollMode = "intoView"
	dispatchScrollModeTop      dispatchScrollMode = "top"
	dispatchScrollModeBottom   dispatchScrollMode = "bottom"
)

func newDispatchPayload(value runtime.Value) (dispatchPayload, error) {
	var payload dispatchPayload

	if value == nil || value == runtime.None {
		return payload, nil
	}

	fields, err := runtime.CastMap(value)
	if err != nil {
		return payload, err
	}

	payload.fields = fields

	return payload, nil
}

func dispatchLookup(ctx context.Context, payload dispatchPayload, key string) (runtime.Value, bool, error) {
	if payload.fields == nil {
		return runtime.None, false, nil
	}

	return payload.fields.Lookup(ctx, runtime.NewString(key))
}

func dispatchRequire(ctx context.Context, payload dispatchPayload, key string) (runtime.Value, error) {
	value, found, err := dispatchLookup(ctx, payload, key)
	if err != nil {
		return runtime.None, err
	}

	if !found || value == runtime.None {
		return runtime.None, runtime.Error(runtime.ErrMissedArgument, key)
	}

	return value, nil
}

func dispatchOptionalString(ctx context.Context, payload dispatchPayload, key, fallback string) (string, error) {
	value, found, err := dispatchLookup(ctx, payload, key)
	if err != nil || !found || value == runtime.None {
		return fallback, err
	}

	str, err := runtime.CastString(value)
	if err != nil {
		return "", err
	}

	return str.String(), nil
}

func dispatchOptionalInt(ctx context.Context, payload dispatchPayload, key string, fallback runtime.Int) (runtime.Int, error) {
	value, found, err := dispatchLookup(ctx, payload, key)
	if err != nil || !found || value == runtime.None {
		return fallback, err
	}

	return runtime.CastInt(value)
}

func dispatchOptionalBool(ctx context.Context, payload dispatchPayload, key string, fallback runtime.Boolean) (runtime.Boolean, error) {
	value, found, err := dispatchLookup(ctx, payload, key)
	if err != nil || !found || value == runtime.None {
		return fallback, err
	}

	return runtime.CastBoolean(value)
}

func dispatchOptionalFloat(ctx context.Context, payload dispatchPayload, key string) (float64, bool, error) {
	value, found, err := dispatchLookup(ctx, payload, key)
	if err != nil || !found || value == runtime.None {
		return 0, false, err
	}

	num, err := runtime.ToFloat(ctx, value)
	if err != nil {
		return 0, false, err
	}

	return float64(num), true, nil
}

func parseDispatchMousePayload(ctx context.Context, event string, value runtime.Value) (input.MouseEventParams, error) {
	payload, err := newDispatchPayload(value)
	if err != nil {
		return input.MouseEventParams{}, err
	}

	button, err := dispatchOptionalString(ctx, payload, "button", "left")
	if err != nil {
		return input.MouseEventParams{}, err
	}

	defaultCount := runtime.NewInt(1)
	if event == drivers.DispatchDoubleClickEvent {
		defaultCount = 2
	}

	count, err := dispatchOptionalInt(ctx, payload, "count", defaultCount)
	if err != nil {
		return input.MouseEventParams{}, err
	}

	params := input.MouseEventParams{
		Button: button,
		Count:  int(count),
	}

	if x, found, err := dispatchOptionalFloat(ctx, payload, "x"); err != nil {
		return input.MouseEventParams{}, err
	} else if found {
		params.X = &x
	}

	if y, found, err := dispatchOptionalFloat(ctx, payload, "y"); err != nil {
		return input.MouseEventParams{}, err
	} else if found {
		params.Y = &y
	}

	return params, nil
}

func parseDispatchKeyboardPayload(ctx context.Context, value runtime.Value) (dispatchKeyboardParams, error) {
	payload, err := newDispatchPayload(value)
	if err != nil {
		return dispatchKeyboardParams{}, err
	}

	count, err := dispatchOptionalInt(ctx, payload, "count", 1)
	if err != nil {
		return dispatchKeyboardParams{}, err
	}

	if count <= 0 {
		count = 1
	}

	if keysValue, found, err := dispatchLookup(ctx, payload, "keys"); err != nil {
		return dispatchKeyboardParams{}, err
	} else if found && keysValue != runtime.None {
		keys, err := sdk.ToSlice(ctx, keysValue, func(_ context.Context, value, _ runtime.Value) (runtime.String, error) {
			return runtime.CastString(value)
		})

		if err != nil {
			return dispatchKeyboardParams{}, err
		}

		return dispatchKeyboardParams{Keys: keys, Count: count}, nil
	}

	keyValue, err := dispatchRequire(ctx, payload, "key")
	if err != nil {
		return dispatchKeyboardParams{}, err
	}

	key, err := runtime.CastString(keyValue)
	if err != nil {
		return dispatchKeyboardParams{}, err
	}

	return dispatchKeyboardParams{
		Keys:  []runtime.String{key},
		Count: count,
	}, nil
}

func parseDispatchKeyPayload(ctx context.Context, value runtime.Value) (string, error) {
	payload, err := newDispatchPayload(value)
	if err != nil {
		return "", err
	}

	keyValue, err := dispatchRequire(ctx, payload, "key")
	if err != nil {
		return "", err
	}

	key, err := runtime.CastString(keyValue)
	if err != nil {
		return "", err
	}

	return key.String(), nil
}

func parseDispatchTypePayload(ctx context.Context, value runtime.Value) (dispatchTypeParams, error) {
	payload, err := newDispatchPayload(value)
	if err != nil {
		return dispatchTypeParams{}, err
	}

	textValue, err := dispatchRequire(ctx, payload, "text")
	if err != nil {
		return dispatchTypeParams{}, err
	}

	text, err := runtime.CastString(textValue)
	if err != nil {
		return dispatchTypeParams{}, err
	}

	delay, err := dispatchOptionalInt(ctx, payload, "delay", runtime.NewInt(drivers.DefaultKeyboardDelay))
	if err != nil {
		return dispatchTypeParams{}, err
	}

	clear, err := dispatchOptionalBool(ctx, payload, "clear", runtime.False)
	if err != nil {
		return dispatchTypeParams{}, err
	}

	return dispatchTypeParams{
		Text:  text.String(),
		Clear: bool(clear),
		Delay: delay,
	}, nil
}

func parseDispatchScrollPayload(ctx context.Context, value runtime.Value) (dispatchScrollParams, error) {
	payload, err := newDispatchPayload(value)
	if err != nil {
		return dispatchScrollParams{}, err
	}

	if payload.fields == nil {
		return dispatchScrollParams{}, runtime.Error(runtime.ErrMissedArgument, "scroll payload")
	}

	options, err := parseDispatchScrollOptions(ctx, payload)
	if err != nil {
		return dispatchScrollParams{}, err
	}

	if intoView, err := dispatchOptionalBool(ctx, payload, "intoView", runtime.False); err != nil {
		return dispatchScrollParams{}, err
	} else if intoView {
		return dispatchScrollParams{Mode: dispatchScrollModeIntoView, Options: options}, nil
	}

	if toValue, found, err := dispatchLookup(ctx, payload, "to"); err != nil {
		return dispatchScrollParams{}, err
	} else if found && toValue != runtime.None {
		params, err := parseDispatchScrollTarget(ctx, toValue, options)
		if err != nil {
			return dispatchScrollParams{}, err
		}

		return params, nil
	}

	if byValue, found, err := dispatchLookup(ctx, payload, "by"); err != nil {
		return dispatchScrollParams{}, err
	} else if found && byValue != runtime.None {
		options, err = parseDispatchScrollCoordinates(ctx, byValue, options)
		if err != nil {
			return dispatchScrollParams{}, err
		}

		return dispatchScrollParams{Mode: dispatchScrollModeBy, Options: options}, nil
	}

	options, found, err := parseDispatchScrollCoordinateFields(ctx, payload, options)
	if err != nil {
		return dispatchScrollParams{}, err
	}

	if !found {
		return dispatchScrollParams{}, runtime.Error(runtime.ErrMissedArgument, "scroll coordinates")
	}

	return dispatchScrollParams{Mode: dispatchScrollModeTo, Options: options}, nil
}

func parseDispatchScrollOptions(ctx context.Context, payload dispatchPayload) (drivers.ScrollOptions, error) {
	var options drivers.ScrollOptions

	if behavior, err := dispatchOptionalString(ctx, payload, "behavior", ""); err != nil {
		return options, err
	} else if behavior != "" {
		options.Behavior = drivers.NewScrollBehavior(behavior)
	}

	if block, err := dispatchOptionalString(ctx, payload, "block", ""); err != nil {
		return options, err
	} else if block != "" {
		options.Block = drivers.NewScrollVerticalAlignment(block)
	}

	if inline, err := dispatchOptionalString(ctx, payload, "inline", ""); err != nil {
		return options, err
	} else if inline != "" {
		options.Inline = drivers.NewScrollHorizontalAlignment(inline)
	}

	return options, nil
}

func parseDispatchScrollTarget(
	ctx context.Context,
	value runtime.Value,
	options drivers.ScrollOptions,
) (dispatchScrollParams, error) {
	if target, err := runtime.CastString(value); err == nil {
		switch target.String() {
		case string(dispatchScrollModeTop):
			return dispatchScrollParams{Mode: dispatchScrollModeTop, Options: options}, nil
		case string(dispatchScrollModeBottom):
			return dispatchScrollParams{Mode: dispatchScrollModeBottom, Options: options}, nil
		default:
			return dispatchScrollParams{}, runtime.Errorf(
				runtime.ErrInvalidOperation,
				"unsupported scroll target %q; supported targets: top, bottom",
				target.String(),
			)
		}
	}

	options, err := parseDispatchScrollCoordinates(ctx, value, options)
	if err != nil {
		return dispatchScrollParams{}, err
	}

	return dispatchScrollParams{Mode: dispatchScrollModeTo, Options: options}, nil
}

func parseDispatchScrollCoordinates(ctx context.Context, value runtime.Value, options drivers.ScrollOptions) (drivers.ScrollOptions, error) {
	payload, err := newDispatchPayload(value)
	if err != nil {
		return options, err
	}

	parsed, found, err := parseDispatchScrollCoordinateFields(ctx, payload, options)
	if err != nil {
		return options, err
	}

	if !found {
		return options, runtime.Error(runtime.ErrMissedArgument, "scroll coordinates")
	}

	return parsed, nil
}

func parseDispatchScrollCoordinateFields(
	ctx context.Context,
	payload dispatchPayload,
	options drivers.ScrollOptions,
) (drivers.ScrollOptions, bool, error) {
	found := false

	if x, ok, err := dispatchOptionalFloat(ctx, payload, "x"); err != nil {
		return options, false, err
	} else if ok {
		options.Left = runtime.NewFloat(x)
		found = true
	}

	if y, ok, err := dispatchOptionalFloat(ctx, payload, "y"); err != nil {
		return options, false, err
	} else if ok {
		options.Top = runtime.NewFloat(y)
		found = true
	}

	return options, found, nil
}

func supportedDispatchEventNames() string {
	return strings.Join(drivers.SupportedDispatchEvents(), ", ")
}
