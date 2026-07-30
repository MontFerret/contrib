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
func MouseMoveXY(ctx context.Context, root, xValue, yValue runtime.Value) (runtime.Value, error) {
	doc, err := drivers.ToDocumentViewportTarget(root)

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(xValue, runtime.TypeInt, runtime.TypeFloat)

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(yValue, runtime.TypeInt, runtime.TypeFloat)

	if err != nil {
		return runtime.None, err
	}

	x, err := runtime.ToFloat(ctx, xValue)

	if err != nil {
		return runtime.None, err
	}

	y, err := runtime.ToFloat(ctx, yValue)

	if err != nil {
		return runtime.None, err
	}

	return doc.MoveMouseByXY(ctx, x, y)
}
