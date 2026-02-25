package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// INPUT types a value to an underlying input element.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} valueOrSelector - CSS selector or a value.
// @param {String} value - Target value.
// @param {Int} [delay] - Target value.
// @return {Boolean} - Returns true if an element was found.
func Input(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 4)

	if err != nil {
		return runtime.False, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.False, err
	}

	delay := runtime.NewInt(drivers.DefaultKeyboardDelay)

	// INPUT(el, value)
	if len(args) == 2 {
		return runtime.True, el.Input(ctx, args[1], delay)
	}

	var selector drivers.QuerySelector
	var value runtime.Value

	// INPUT(el, valueOrSelector, valueOrOpts)
	if len(args) == 3 {
		switch v := args[2].(type) {
		// INPUT(el, value, delay)
		case runtime.Int, runtime.Float:
			value = args[1]
			delay, err = runtime.ToInt(ctx, v)

			if err != nil {
				return runtime.False, err
			}

			return runtime.True, el.Input(ctx, value, delay)
		default:
			// INPUT(el, selector, value)
			selector, err = drivers.ToQuerySelector(args[1])

			if err != nil {
				return runtime.None, err
			}

			value = args[2]
		}
	} else {
		// INPUT(el, selector, value, delay)
		if err := runtime.ValidateType(args[3], runtime.TypeInt); err != nil {
			return runtime.False, err
		}

		selector, err = drivers.ToQuerySelector(args[1])

		if err != nil {
			return runtime.None, err
		}

		value = args[2]
		delay, err = runtime.ToInt(ctx, args[3])

		if err != nil {
			return runtime.False, err
		}
	}

	exists, err := el.ExistsBySelector(ctx, selector)

	if err != nil {
		return runtime.False, err
	}

	if !exists {
		return runtime.False, nil
	}

	return runtime.True, el.InputBySelector(ctx, selector, value, delay)
}
