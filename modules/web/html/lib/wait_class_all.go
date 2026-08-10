package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// WaitClassAll waits for a class on all matching elements.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param class {String} Class name.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the class appears on every match.
func WaitClassAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitClassAllWhen(ctx, args, drivers.WaitEventPresence)
}

// WaitNoClassAll waits for a class to disappear from all matching elements.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @param class {String} Class name.
// @param timeout {Int?} Wait timeout in milliseconds.
// @return {Boolean} True when the class disappears from every match.
func WaitNoClassAll(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return waitClassAllWhen(ctx, args, drivers.WaitEventAbsence)
}

func waitClassAllWhen(ctx context.Context, args []runtime.Value, when drivers.WaitEvent) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 3, 4)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootWaitTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	// selector
	selector, err := drivers.ToQuerySelector(ctx, args[1])

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

	return runtime.True, target.WaitForClassBySelectorAll(ctx, selector, class, when)
}
