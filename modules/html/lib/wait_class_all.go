package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WAIT_CLASS_ALL waits for a class to appear on all matched elements.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - String of CSS selector.
// @param {String} class - String of target CSS class.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitClassAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitClassAllWhen(ctx, args, drivers.WaitEventPresence)
}

// WAIT_NO_CLASS_ALL waits for a class to disappear on all matched elements.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - String of CSS selector.
// @param {String} class - String of target CSS class.
// @param {Int} [timeout=5000] - Wait timeout.
func WaitNoClassAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitClassAllWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitClassAllWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 4)

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

	// class
	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	class := args[2].(runtime.String)
	timeout := runtime.NewInt(drivers.DefaultWaitTimeout)

	if len(args) == 4 {
		err = runtime.ValidateType(args[3], runtime.TypeInt)

		if err != nil {
			return runtime.None, err
		}

		timeout = args[3].(runtime.Int)
	}

	ctx, fn := waitTimeout(ctx, timeout)
	defer fn()

	return runtime.True, el.WaitForClassBySelectorAll(ctx, selector, class, when)
}
