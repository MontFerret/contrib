package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WaitElement waits for an element to appear.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the element appears.
func WaitElement(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitElementWhen(ctx, args, drivers.WaitEventPresence)
}

// WaitNoElement waits for an element to disappear.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the element disappears.
func WaitNoElement(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitElementWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitElementWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootWaitTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

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

	return runtime.True, target.WaitForElement(ctx, selector, when)
}
