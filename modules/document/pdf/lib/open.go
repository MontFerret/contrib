package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// openWithOptions opens an existing PDF through the configured filesystem.
//
// The path is resolved through the host's Ferret filesystem policy.
//
// @param path {String} Sandboxed PDF path.
// @return {PDFDocument} Open PDF document handle.
func openWithOptions(options core.OpenOptions) func(context.Context, runtime.Value) (runtime.Value, error) {
	return func(ctx context.Context, pathValue runtime.Value) (runtime.Value, error) {
		path, err := requireString(pathValue, "OPEN", "path")
		if err != nil {
			return runtime.None, err
		}

		return core.Open(ctx, path, options)
	}
}
