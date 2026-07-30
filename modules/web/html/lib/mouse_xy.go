package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// MouseMoveXY moves the mouse to given viewport coordinates.
// @param {HTMLPage | HTMLDocument} document - HTML page or document.
// @param {Int|Float} x - X coordinate.
// @param {Int|Float} y - Y coordinate.
// @return {Boolean} - True if the mouse was moved, otherwise false.
func MouseMoveXY(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 3)

	if err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToDocumentViewportTarget(args[0])

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

	x, err := runtime.ToFloat(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	y, err := runtime.ToFloat(ctx, args[2])

	if err != nil {
		return runtime.None, err
	}

	return doc.MoveMouseByXY(ctx, x, y)
}
