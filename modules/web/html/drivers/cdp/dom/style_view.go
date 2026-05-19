package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type styleView struct {
	*elementMapView
}

func newStyleView(ctx context.Context, styles *elementStyles) (*styleView, error) {
	snapshot, err := styles.GetStyles(ctx)
	if err != nil {
		return nil, err
	}

	return &styleView{
		elementMapView: newElementMapView(
			snapshot,
			func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
				name := runtime.ToString(key)
				if value == runtime.None {
					return runtime.None, true, styles.RemoveStyle(ctx, name)
				}

				next := runtime.ToString(value)

				return next, false, styles.SetStyle(ctx, name, next)
			},
			func(ctx context.Context, key runtime.Value) error {
				return styles.RemoveStyle(ctx, runtime.ToString(key))
			},
		),
	}, nil
}
