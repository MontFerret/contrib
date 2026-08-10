package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// ElementExists reports whether an element matches a selector.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String} Element selector.
// @return {Boolean} Whether a matching element exists.
func ElementExists(ctx context.Context, root, selectorValue runtime.Value) (runtime.Value, error) {
	el, selector, err := queryArgs(ctx, root, selectorValue)

	if err != nil {
		return runtime.None, err
	}

	return el.ExistsBySelector(ctx, selector)
}
