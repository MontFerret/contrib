package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WAIT_STYLE_ALL waits until a target style value appears on all matched elements with a given value.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} styleNameOrSelector - Style name or CSS selector.
// @param {String | Any} valueOrStyleName - Style value or name.
// @param {Any | Int} [valueOrTimeout] - Style value or wait timeout.
// @param {Int} [timeout=5000] - Timeout.
func WaitStyleAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitStyleAllWhen(ctx, args, drivers.WaitEventPresence)
}

// WAIT_NO_STYLE_ALL waits until a target style value disappears on all matched elements with a given value.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} styleNameOrSelector - Style name or CSS selector.
// @param {String | Any} valueOrStyleName - Style value or name.
// @param {Any | Int} [valueOrTimeout] - Style value or wait timeout.
// @param {Int} [timeout=5000] - Timeout.
func WaitNoStyleAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitStyleAllWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitStyleAllWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 4, 5)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

	// selector
	selector, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return runtime.None, err
	}

	// attr name
	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	name := args[2].(runtime.String)
	value := args[3]
	timeout := runtime.NewInt(drivers.DefaultWaitTimeout)

	if len(args) == 5 {
		err = runtime.ValidateType(args[4], runtime.TypeInt)

		if err != nil {
			return runtime.None, err
		}

		timeout = args[4].(runtime.Int)
	}

	ctx, fn := waitTimeout(ctx, timeout)
	defer fn()

	return runtime.True, el.WaitForStyleBySelectorAll(ctx, selector, name, value, when)
}
