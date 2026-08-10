package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Decode decodes a single YAML document.
//
// @param data {String|Binary} YAML content.
// @return {Any} Decoded value.
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
