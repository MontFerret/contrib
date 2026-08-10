package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// GetInnerHTML returns HTML from a root or selected element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param selector {String?} Element selector.
// @return {String} Inner HTML, or an empty string when no element matches.
func GetInnerHTML(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.EmptyString, err
	}

	target, err := toRootContentTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 1 {
		return target.GetInnerHTML(ctx)
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	return target.GetInnerHTMLBySelector(ctx, selector)
}
