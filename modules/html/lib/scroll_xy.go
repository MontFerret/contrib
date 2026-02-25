package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// SCROLL scrolls by given coordinates.
// @param {HTMLDocument} document - HTML document.
// @param {Int | Float} x - X coordinate.
// @param {Int | Float} y - Y coordinate.
// @param {Object} [params] - Scroll params.
// @param {String} [params.behavior="instant"] - Scroll behavior
// @param {String} [params.block="center"] - Scroll vertical alignment.
// @param {String} [params.inline="center"] - Scroll horizontal alignment.
func ScrollXY(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 3, 4); err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToDocument(args[0])

	if err != nil {
		return runtime.None, err
	}

	if err = runtime.ValidateType(args[1], runtime.TypeInt, runtime.TypeFloat); err != nil {
		return runtime.None, err
	}

	if err = runtime.ValidateType(args[2], runtime.TypeInt, runtime.TypeFloat); err != nil {
		return runtime.None, err
	}

	x, err := runtime.ToFloat(ctx, args[1])
	if err != nil {
		return runtime.None, err
	}

	y, err := runtime.ToFloat(ctx, args[2])
	if err != nil {
		return runtime.None, err
	}

	var opts drivers.ScrollOptions
	opts.Left = x
	opts.Top = y

	if len(args) > 3 {
		opts, err = toScrollOptions(args[3])

		if err != nil {
			return runtime.None, err
		}

		opts.Left = x
		opts.Top = y
	}

	return runtime.True, doc.Scroll(ctx, opts)
}
