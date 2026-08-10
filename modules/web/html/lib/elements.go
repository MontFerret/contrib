package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Elements returns all elements matching a selector.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @return {Array<HTMLElement>} Matching elements.
func Elements(ctx context.Context, root, selectorValue runtime.Value) (runtime.Value, error) {
	el, selector, err := queryArgs(ctx, root, selectorValue)

	if err != nil {
		return runtime.None, err
	}

	return el.QuerySelectorAll(ctx, selector)
}
