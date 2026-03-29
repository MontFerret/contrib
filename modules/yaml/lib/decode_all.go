package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeAll decodes all YAML documents from a YAML stream.
// @param {String|Binary} data - YAML content.
// @return {Any[]} - Decoded YAML values, one per document.
func DecodeAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	content, err := core.ResolveContent(args[0])
	if err != nil {
		return nil, err
	}

	return core.DecodeAll(ctx, content)
}
