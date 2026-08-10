package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// SetInnerHTML sets HTML on a root or selected element.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param htmlOrSelector {String} HTML content or element selector.
// @param html {String?} HTML content when a selector is supplied.
// @return {None} No value.
func SetInnerHTML(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, 3)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootContentTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	if len(args) == 2 {
		err := runtime.ValidateType(args[1], runtime.TypeString)

		if err != nil {
			return runtime.None, err
		}

		return runtime.None, target.SetInnerHTML(ctx, runtime.ToString(args[1]))
	}

	selector, err := drivers.ToQuerySelector(ctx, args[1])

	if err != nil {
		return runtime.None, err
	}

	err = runtime.ValidateType(args[2], runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	innerHTML := runtime.ToString(args[2])

	return runtime.None, target.SetInnerHTMLBySelector(ctx, selector, innerHTML)
}
