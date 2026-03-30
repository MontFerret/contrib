package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const (
	DecodeDateTimeString = "string"
	DecodeDateTimeNative = "native"

	EncodeDateTimeRFC3339  = "rfc3339"
	EncodeDateTimePreserve = "preserve"
)

// DecodeOptions configures TOML decoding behavior.
type DecodeOptions struct {
	DateTime string
	Strict   bool
}

// EncodeOptions configures TOML encoding behavior.
type EncodeOptions struct {
	DateTime string
	SortKeys bool
}

// DefaultDecodeOptions returns the default TOML decode options.
func DefaultDecodeOptions() DecodeOptions {
	return DecodeOptions{
		DateTime: DecodeDateTimeString,
		Strict:   true,
	}
}

// DefaultEncodeOptions returns the default TOML encode options.
func DefaultEncodeOptions() EncodeOptions {
	return EncodeOptions{
		SortKeys: false,
		DateTime: EncodeDateTimeRFC3339,
	}
}

// ParseDecodeOptions validates and decodes TOML decode options.
func ParseDecodeOptions(ctx context.Context, input runtime.Value) (DecodeOptions, error) {
	opts := DefaultDecodeOptions()

	m, err := runtime.CastMap(input)
	if err != nil {
		return opts, err
	}

	if err := m.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, err := optionKey(key)
		if err != nil {
			return false, err
		}

		switch name {
		case "datetime":
			raw, ok := value.(runtime.String)
			if !ok {
				return false, newError(`decode option "datetime" must be a string`)
			}

			switch raw.String() {
			case DecodeDateTimeString, DecodeDateTimeNative:
				opts.DateTime = raw.String()
			default:
				return false, newErrorf(`decode option "datetime" must be %q or %q`, DecodeDateTimeString, DecodeDateTimeNative)
			}
		case "strict":
			raw, ok := value.(runtime.Boolean)
			if !ok {
				return false, newError(`decode option "strict" must be a boolean`)
			}

			opts.Strict = bool(raw)
		default:
			return false, newErrorf("unknown decode option %q", name)
		}

		return true, nil
	}); err != nil {
		return opts, err
	}

	return opts, nil
}

// ParseEncodeOptions validates and decodes TOML encode options.
func ParseEncodeOptions(ctx context.Context, input runtime.Value) (EncodeOptions, error) {
	opts := DefaultEncodeOptions()

	m, err := runtime.CastMap(input)
	if err != nil {
		return opts, err
	}

	if err := m.ForEach(ctx, func(_ context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		name, err := optionKey(key)
		if err != nil {
			return false, err
		}

		switch name {
		case "sortKeys":
			raw, ok := value.(runtime.Boolean)
			if !ok {
				return false, newError(`encode option "sortKeys" must be a boolean`)
			}

			opts.SortKeys = bool(raw)
		case "datetime":
			raw, ok := value.(runtime.String)
			if !ok {
				return false, newError(`encode option "datetime" must be a string`)
			}

			switch raw.String() {
			case EncodeDateTimeRFC3339, EncodeDateTimePreserve:
				opts.DateTime = raw.String()
			default:
				return false, newErrorf(`encode option "datetime" must be %q or %q`, EncodeDateTimeRFC3339, EncodeDateTimePreserve)
			}
		default:
			return false, newErrorf("unknown encode option %q", name)
		}

		return true, nil
	}); err != nil {
		return opts, err
	}

	return opts, nil
}

func optionKey(key runtime.Value) (string, error) {
	str, ok := key.(runtime.String)
	if !ok {
		return "", newError("option keys must be strings")
	}

	return str.String(), nil
}
