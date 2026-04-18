package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Element finds an element by a given CSS selector.
// Returns NONE if element not found.
// @param {HTMLPage | HTMLDocument | HTMLElement} node - Target html node.
// @param {String} selector - CSS selector.
// @return {HTMLElement} - A matched HTML element
func Element(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	el, selector, err := queryArgs(args)

	if err != nil {
		return runtime.None, err
	}

	return el.QuerySelector(ctx, selector)
}

func queryArgs(args []runtime.Value) (drivers.QueryTarget, drivers.QuerySelector, error) {
	if err := runtime.ValidateArgs(args, 2, 2); err != nil {
		return nil, drivers.QuerySelector{}, err
	}

	target, err := drivers.ToQueryTarget(args[0])

	if err != nil {
		return nil, drivers.QuerySelector{}, err
	}

	qs, err := drivers.ToQuerySelector(args[1])

	if err != nil {
		return nil, drivers.QuerySelector{}, err
	}

	return target, qs, nil
}
