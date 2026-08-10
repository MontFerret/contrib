package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Navigate loads a URL in a page and waits for navigation to finish.
//
// @param page {HTMLPage} Target page.
// @param url {String} Destination URL.
// @param timeout {Int?} Navigation timeout in milliseconds.
// @return {Boolean} True when navigation succeeds.
func Navigate(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	target, err := drivers.ToPageNavigationTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(args[1], runtime.TypeString)

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

	return runtime.True, target.Navigate(ctx, args[1].(runtime.String))
}
