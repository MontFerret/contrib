package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Encode serializes a Ferret runtime value into YAML text.
// @param {Any} value - Ferret runtime value.
// @return {String} - YAML text.
func Encode(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	text, err := core.Encode(ctx, args[0])
	if err != nil {
		return nil, err
	}

	return runtime.NewString(text), nil
}
