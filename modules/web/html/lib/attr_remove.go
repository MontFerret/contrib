package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// AttributeRemove removes attributes from an HTML root.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param names {String...} Attribute names to remove.
// @return {None} No value.
func AttributeRemove(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootAttributeTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	attrs := args[1:]
	attrsStr := make([]runtime.String, 0, len(attrs))

	for _, attr := range attrs {
		str, ok := attr.(runtime.String)

		if !ok {
			return runtime.None, runtime.TypeErrorOf(attr, runtime.TypeString)
		}

		attrsStr = append(attrsStr, str)
	}

	return runtime.None, target.RemoveAttribute(ctx, attrsStr...)
}
