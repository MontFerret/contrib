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
func Element(ctx context.Context, root, selectorValue runtime.Value) (runtime.Value, error) {
	el, selector, err := queryArgs(ctx, root, selectorValue)

	if err != nil {
		return runtime.None, err
	}

	return el.QuerySelector(ctx, selector)
}

func queryArgs(ctx context.Context, root, selectorValue runtime.Value) (drivers.QueryTarget, drivers.QuerySelector, error) {
	target, err := drivers.ToQueryTarget(root)

	if err != nil {
		return nil, drivers.QuerySelector{}, err
	}

	qs, err := drivers.ToQuerySelector(ctx, selectorValue)

	if err != nil {
		return nil, drivers.QuerySelector{}, err
	}

	return target, qs, nil
}
