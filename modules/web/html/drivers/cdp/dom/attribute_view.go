package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type attributeView struct {
	*elementMapView
}

func newAttributeView(ctx context.Context, el *HTMLElement) (*attributeView, error) {
	snapshot, err := el.GetAttributes(ctx)
	if err != nil {
		return nil, err
	}

	return &attributeView{
		elementMapView: newElementMapView(
			snapshot,
			func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
				name := runtime.ToString(key)
				if value == runtime.None {
					return runtime.None, true, el.RemoveAttribute(ctx, name)
				}

				next := runtime.ToString(value)

				return next, false, el.SetAttribute(ctx, name, next)
			},
			func(ctx context.Context, key runtime.Value) error {
				return el.RemoveAttribute(ctx, runtime.ToString(key))
			},
		),
	}, nil
}
