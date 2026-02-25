package common

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func ToRuntimeStringSlice(ctx context.Context, input runtime.Value) ([]runtime.String, error) {
	return sdk.ToSlice(ctx, input, func(ctx context.Context, value, key runtime.Value) (runtime.String, error) {
		return runtime.String(key.String()), nil
	})
}
