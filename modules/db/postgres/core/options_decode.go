package core

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DecodeOpenOptions decodes a Ferret options object into OpenOptions.
func DecodeOpenOptions(value runtime.Value) (OpenOptions, error) {
	optsMap, err := runtime.Cast[runtime.Map](value)
	if err != nil {
		return OpenOptions{}, OperationError("OPEN", err)
	}

	var opts OpenOptions
	if err := sdk.Decode(optsMap, &opts); err != nil {
		return OpenOptions{}, OperationError("OPEN", err)
	}

	return opts, nil
}
