package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// STYLE_SET sets or updates a single or more style attribute value of a given element.
// @param {HTMLElement} element - Target html element.
// @param {String | Object} nameOrObj - Style name or an object representing a key-value pair of attributes.
// @param {String} value - If a second parameter is a string value, this parameter represent a style value.
func StyleSet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	switch arg1 := args[1].(type) {
	case runtime.String:
		// STYLE_SET(el, name, value)
		err = runtime.ValidateArgs(args, 3, 3)

		if err != nil {
			return runtime.None, nil
		}

		return runtime.None, el.SetStyle(ctx, arg1, runtime.NewString(args[2].String()))
	case runtime.Map:
		// STYLE_SET(el, values)
		return runtime.None, el.SetStyles(ctx, arg1)
	default:
		return runtime.None, runtime.TypeError(runtime.TypeOf(arg1), runtime.TypeString, runtime.TypeObject)
	}
}
