package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// NAVIGATE navigates a given page to a new resource.
// The operation blocks the execution until the page gets loaded.
// Which means there is no need in WAIT_NAVIGATION function.
// @param {HTMLPage} page - Target page.
// @param {String} url - Target url to navigate.
// @param {Int} [timeout=5000] - Navigation timeout.
func Navigate(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	page, err := drivers.ToPage(args[0])

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

	return runtime.True, page.Navigate(ctx, args[1].(runtime.String))
}
