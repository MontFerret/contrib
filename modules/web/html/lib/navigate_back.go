package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// NavigateBack moves backward through page history.
//
// @param page {HTMLPage} Target page.
// @param skip {Int?} Number of history entries to skip.
// @param timeout {Int?} Navigation timeout in milliseconds.
// @return {Boolean} Whether history existed and navigation succeeded.
func NavigateBack(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 3)

	if err != nil {
		return runtime.False, err
	}

	target, err := drivers.ToPageNavigationTarget(args[0])

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

	return target.NavigateBack(ctx, skip)
}
