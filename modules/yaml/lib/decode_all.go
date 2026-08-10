package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/yaml/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeAll decodes all YAML documents from a YAML stream.
//
// @param data {String|Binary} YAML content.
// @return {Array<Any>} Decoded values in document order.
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
