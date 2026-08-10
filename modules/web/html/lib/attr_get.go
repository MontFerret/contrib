package lib

import (
	"context"
	"errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// AttributeGet returns selected attributes from an HTML root.
//
// @param root {HTMLPage|HTMLDocument|HTMLElement} HTML root.
// @param names {String...} Attribute names.
// @return {Object} Existing attribute values keyed by name.
func AttributeGet(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 2, runtime.MaxArgs)

	if err != nil {
		return runtime.None, err
	}

	target, err := toRootAttributeTarget(args[0])

	if err != nil {
		return runtime.None, err
	}

	names := args[1:]
	result := runtime.NewObject()
	attrs, err := target.GetAttributes(ctx)

	if err != nil {
		return runtime.None, err
	}

	for _, n := range names {
		name := runtime.NewString(n.String())
		val, err := attrs.Get(ctx, name)

		if err != nil && !errors.Is(err, drivers.ErrNotFound) && !errors.Is(err, runtime.ErrNotFound) {
			return runtime.None, err
		}

		if err == nil {
			_ = result.Set(ctx, name, val)
		}
	}

	return result, nil
}
