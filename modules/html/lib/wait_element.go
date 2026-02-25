package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WAIT_ELEMENT waits for element to appear in the DOM.
// Stops the execution until it finds an element or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - Target element's selector.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitElement(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitElementWhen(ctx, args, drivers.WaitEventPresence)
}

// WAIT_NO_ELEMENT waits for element to disappear in the DOM.
// Stops the execution until it does not find an element or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - Target element's selector.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitNoElement(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitElementWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitElementWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

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

	timeout := runtime.NewInt(drivers.DefaultWaitTimeout)

	if len(args) > 2 {
		err = runtime.ValidateType(args[2], runtime.TypeInt)

		if err != nil {
			return runtime.None, err
		}

		timeout = args[2].(runtime.Int)
	}

	ctx, fn := waitTimeout(ctx, timeout)
	defer fn()

	return runtime.True, el.WaitForElement(ctx, selector, when)
}
