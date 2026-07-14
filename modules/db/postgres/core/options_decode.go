package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeOpenOptions decodes a Ferret options object into OpenOptions.
func DecodeOpenOptions(ctx context.Context, value runtime.Value) (OpenOptions, error) {
	optsMap, err := runtime.Cast[runtime.Map](value)
	if err != nil {
		return OpenOptions{}, OperationError("OPEN", err)
	}

	var opts OpenOptions
	if err := sdk.Decode(ctx, optsMap, &opts, sdk.DisallowUnknownFields()); err != nil {
		return OpenOptions{}, OperationError("OPEN", err)
	}

	return opts, nil
}
