package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// STYLE_GET gets single or more style attribute value(s) of a given element.
// @param {HTMLElement} element - Target html element.
// @param {String, repeated} names - Style name(s).
// @return {Object} - Collection of key-value pairs of style runtime.
func StyleGet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	names := args[1:]
	result := runtime.NewObject()

	for _, n := range names {
		name := runtime.NewString(n.String())
		val, err := el.GetStyle(ctx, name)

		if err != nil {
			return runtime.None, err
		}

		if val != runtime.None {
			if err := result.Set(ctx, name, val); err != nil {
				return runtime.None, err
			}
		}
	}

	return result, nil
}
