package core

import (
	"context"
	"errors"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func decodeOptionObject[T any](
	ctx context.Context,
	value runtime.Value,
	label string,
	options ...sdk.DecodeOption,
) (T, error) {
	var zero T

	if value == nil || runtime.TypeNone.Is(value) {
		return zero, nil
	}

	return decodeValue[T](ctx, value, label, runtime.TypeMap, options...)
}

// snapshotOptionObject preserves source and cancellation error precedence while
// avoiding a second traversal of callback-backed maps during mode-specific decoding.
func snapshotOptionObject(ctx context.Context, value runtime.Value, label string) (runtime.Value, error) {
	if value == nil || runtime.TypeNone.Is(value) {
		return value, nil
	}

	fields, err := decodeValue[map[string]runtime.Value](ctx, value, label, runtime.TypeMap)
	if err != nil {
		return nil, err
	}

	return runtime.NewObjectWith(fields), nil
}

func decodeList[T any](ctx context.Context, value runtime.Value, label string) (T, error) {
	return decodeValue[T](ctx, value, label, runtime.TypeList)
}

func decodeValue[T any](
	ctx context.Context,
	value runtime.Value,
	label string,
	expected runtime.Type,
	options ...sdk.DecodeOption,
) (T, error) {
	decodeOptions := []sdk.DecodeOption{
		sdk.RequireType(expected),
		sdk.DisallowUnknownFields(),
		sdk.DisallowNoneValues(),
	}
	decodeOptions = append(decodeOptions, options...)
	decoded, err := sdk.DecodeValue[T](ctx, value, decodeOptions...)

	if err != nil {
		var zero T

		return zero, normalizeDecodeError(label, err)
	}

	return decoded, nil
}

func normalizeDecodeError(label string, err error) error {
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	message := label + " are invalid"
	if decodeErr, ok := errors.AsType[*sdk.DecodeError](err); ok && decodeErr.SafeToExpose() {
		message += ": " + decodeErr.Error()
	}

	return NewError(ErrInvalidOptions, message)
}
