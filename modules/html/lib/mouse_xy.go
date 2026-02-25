package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// MOUSE moves mouse by given coordinates.
// @param {HTMLDocument} document - HTML document.
// @param {Int|Float} x - X coordinate.
// @param {Int|Float} y - Y coordinate.
func MouseMoveXY(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 3)

	if err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToDocument(args[0])

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(args[1], runtime.TypeInt, runtime.TypeFloat)

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(args[2], runtime.TypeInt, runtime.TypeFloat)

	if err != nil {
		return runtime.None, err
	}

	x, err := runtime.ToFloat(ctx, args[0])

	if err != nil {
		return runtime.None, err
	}

	y, err := runtime.ToFloat(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	return runtime.None, doc.MoveMouseByXY(ctx, x, y)
}
