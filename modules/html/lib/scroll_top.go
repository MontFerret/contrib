package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// SCROLL_TOP scrolls the document's window to its top.
// @param {HTMLDocument} document - HTML document.
// @param {Int | Float} x - X coordinate.
// @param {Int | Float} y - Y coordinate.
// @param {Object} [params] - Scroll params.
// @param {String} [params.behavior="instant"] - Scroll behavior
// @param {String} [params.block="center"] - Scroll vertical alignment.
// @param {String} [params.inline="center"] - Scroll horizontal alignment.
func ScrollTop(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToDocument(args[0])

	if err != nil {
		return runtime.None, err
	}

	var opts drivers.ScrollOptions

	if len(args) > 1 {
		opts, err = toScrollOptions(args[1])

		if err != nil {
			return runtime.None, err
		}
	}

	return runtime.True, doc.ScrollTop(ctx, opts)
}
