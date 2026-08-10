package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// StyleGet returns selected style values from an element.
//
// @param element {HTMLElement} Target element.
// @param names {String...} Style names.
// @return {Object} Existing style values keyed by name.
func StyleGet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

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

	names := args[1:]
	result := runtime.NewObject()

	for _, n := range names {
		name := runtime.NewString(n.String())
		val, err := styles.GetStyle(ctx, name)

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
