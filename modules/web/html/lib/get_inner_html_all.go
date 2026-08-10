package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// GetInnerHTMLAll returns HTML from every matching element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @return {Array<String>} Inner HTML values in match order.
func GetInnerHTMLAll(ctx context.Context, root, selectorValue runtime.Value) (runtime.Value, error) {
	target, err := toRootContentTarget(root)

	if err != nil {
		return runtime.None, err
	}

	selector, err := drivers.ToQuerySelector(ctx, selectorValue)

	if err != nil {
		return runtime.None, err
	}

	return target.GetInnerHTMLBySelectorAll(ctx, selector)
}
