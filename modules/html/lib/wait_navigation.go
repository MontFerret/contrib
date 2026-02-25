package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type WaitNavigationParams struct {
	TargetURL runtime.String
	Timeout   runtime.Int
	Frame     drivers.HTMLDocument
}

// WAIT_NAVIGATION waits for a given page to navigate to a new url.
// Stops the execution until the navigation ends or operation times out.
// @param {HTMLPage} page - Target page.
// @param {Int} [timeout=5000] - Navigation timeout.
// @param {Object} [params=None] - Navigation parameters.
// @param {Int} [params.timeout=5000] - Navigation timeout.
// @param {String} [params.target] - Navigation target url.
// @param {HTMLDocument} [params.frame] - Navigation frame.
func WaitNavigation(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	doc, err := drivers.ToPage(args[0])

	if err != nil {
		return runtime.None, err
	}

	var params WaitNavigationParams

	if len(args) > 1 {
		p, err := parseWaitNavigationParams(args[1])

		if err != nil {
			return runtime.None, err
		}

		params = p
	} else {
		params = defaultWaitNavigationParams()
	}

	ctx, fn := waitTimeout(ctx, params.Timeout)
	defer fn()

	if params.Frame == nil {
		return runtime.True, doc.WaitForNavigation(ctx, params.TargetURL)
	}

	return runtime.True, doc.WaitForFrameNavigation(ctx, params.Frame, params.TargetURL)
}

func parseWaitNavigationParams(arg runtime.Value) (WaitNavigationParams, error) {
	params := defaultWaitNavigationParams()

	if err := runtime.ValidateType(arg, runtime.TypeInt, runtime.TypeObject); err != nil {
		return params, err
	}

	switch argv := arg.(type) {
	case runtime.Int:
		params.Timeout = argv
	case runtime.Map:
		if err := sdk.Decode(argv, &params); err != nil {
			return params, err
		}
	}

	return params, nil
}

func defaultWaitNavigationParams() WaitNavigationParams {
	return WaitNavigationParams{
		TargetURL: "",
		Timeout:   runtime.NewInt(drivers.DefaultWaitTimeout),
	}
}
