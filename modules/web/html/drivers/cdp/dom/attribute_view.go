package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type attributeView struct {
	*elementMapView
}

func newAttributeView(ctx context.Context, attrs *elementAttributes) (*attributeView, error) {
	snapshot, err := attrs.GetAttributes(ctx)
	if err != nil {
		return nil, err
	}

	return &attributeView{
		elementMapView: newElementMapView(
			snapshot,
			func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
				name := runtime.ToString(key)
				if value == runtime.None {
					return runtime.None, true, attrs.RemoveAttribute(ctx, name)
				}

				next := runtime.ToString(value)

				return next, false, attrs.SetAttribute(ctx, name, next)
			},
			func(ctx context.Context, key runtime.Value) error {
				return attrs.RemoveAttribute(ctx, runtime.ToString(key))
			},
		),
	}, nil
}
