package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// StyleSet sets one or more style values on an element.
//
// @param element {HTMLElement} Target element.
// @param nameOrStyles {String|Object} Style name or style map.
// @param value {Any?} Style value when a name is supplied.
// @return {None} No value.
func StyleSet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	styles, err := drivers.ToStyleTarget(el)

	if err != nil {
		return runtime.None, err
	}

	switch arg1 := args[1].(type) {
	case runtime.String:
		// STYLE_SET(el, name, value)
		err = runtime.ValidateArgs(args, 3, 3)

		if err != nil {
			return runtime.None, err
		}

		return runtime.None, styles.SetStyle(ctx, arg1, runtime.NewString(args[2].String()))
	case runtime.Map:
		// STYLE_SET(el, values)
		return runtime.None, styles.SetStyles(ctx, arg1)
	default:
		return runtime.None, runtime.TypeError(runtime.TypeOf(arg1), runtime.TypeString, runtime.TypeObject)
	}
}
