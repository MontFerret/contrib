package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Decode decodes a single YAML document into a Ferret runtime value.
// @param {String|Binary} data - YAML content.
// @return {Any} - Decoded YAML value.
func Decode(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	content, err := core.ResolveContent(args[0])
	if err != nil {
		return nil, err
	}

	return core.Decode(ctx, content)
}
