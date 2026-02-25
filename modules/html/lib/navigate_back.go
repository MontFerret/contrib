package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// NAVIGATE_BACK navigates a given page back within its navigation history.
// The operation blocks the execution until the page gets loaded.
// If the history is empty, the function returns FALSE.
// @param {HTMLPage} page - Target page.
// @param {Int} [entry=1] - An integer value indicating how many pages to skip.
// @param {Int} [timeout=5000] - Navigation timeout.
// @return {Boolean} - True if history exists and the operation succeeded, otherwise false.
func NavigateBack(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 3)

	if err != nil {
		return runtime.False, err
	}

	page, err := drivers.ToPage(args[0])

	if err != nil {
		return runtime.None, err
	}

	skip := runtime.NewInt(1)
	timeout := runtime.NewInt(drivers.DefaultWaitTimeout)

	if len(args) > 1 {
		err = runtime.ValidateType(args[1], runtime.TypeInt)

		if err != nil {
			return runtime.None, err
		}

		skip = args[1].(runtime.Int)
	}

	if len(args) > 2 {
		err = runtime.ValidateType(args[2], runtime.TypeInt)

		if err != nil {
			return runtime.None, err
		}

		timeout = args[2].(runtime.Int)
	}

	ctx, fn := waitTimeout(ctx, timeout)
	defer fn()

	return page.NavigateBack(ctx, skip)
}
