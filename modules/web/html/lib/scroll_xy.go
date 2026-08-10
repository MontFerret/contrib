package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ScrollXY scrolls a page or document to absolute coordinates.
//
// @param root {HTMLPage|HTMLDocument} Page or document.
// @param x {Number} Horizontal coordinate.
// @param y {Number} Vertical coordinate.
// @param options {Object?} Scroll behavior and alignment options.
// @return {Boolean} Whether scrolling was initiated.
func ScrollXY(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 3, 4); err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToDocumentViewportTarget(args[0])

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
		opts, err = toScrollOptions(ctx, args[3])

		if err != nil {
			return runtime.None, err
		}

		opts.Left = x
		opts.Top = y
	}

	return doc.Scroll(ctx, opts)
}
