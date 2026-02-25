package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WAIT_ATTR_ALL waits for an attribute to appear on all matched elements with a given value.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - String of CSS selector.
// @param {String} class - String of target CSS class.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitAttributeAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitAttributeAllWhen(ctx, args, drivers.WaitEventPresence)
}

// WAIT_NO_ATTR_ALL waits for an attribute to disappear on all matched elements by a given value.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - String of CSS selector.
// @param {String} class - String of target CSS class.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitNoAttributeAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitAttributeAllWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitAttributeAllWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 4, 5)

	if err != nil {
		return runtime.None, err
	}

	el, err := drivers.ToElement(args[0])

	if err != nil {
		return runtime.None, err
	}

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

	return runtime.True, el.WaitForAttributeBySelectorAll(ctx, selector, name, value, when)
}
