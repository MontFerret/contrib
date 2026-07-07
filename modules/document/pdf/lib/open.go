package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func openWithOptions(options core.OpenOptions) func(context.Context, runtime.Value) (runtime.Value, error) {
	return func(ctx context.Context, pathValue runtime.Value) (runtime.Value, error) {
		path, err := requireString(pathValue, "OPEN", "path")
		if err != nil {
			return runtime.None, err
		}

		return core.Open(ctx, path, options)
	}
}
