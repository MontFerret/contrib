package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ScrollBottom scrolls a page or document to the bottom.
//
// @param root {HTMLPage|HTMLDocument} Page or document.
// @param options {Object?} Scroll behavior and alignment options.
// @return {Boolean} Whether scrolling was initiated.
func ScrollBottom(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToDocumentViewportTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	var opts drivers.ScrollOptions

	if len(args) > 1 {
		opts, err = toScrollOptions(ctx, args[1])

		if err != nil {
			return runtime.None, err
		}
	}

	return doc.ScrollBottom(ctx, opts)
}
