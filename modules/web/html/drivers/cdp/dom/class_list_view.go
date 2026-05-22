package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type classListView struct {
	*elementMapView
}

func newClassListView(ctx context.Context, classes *elementClasses) (*classListView, error) {
	snapshot, err := classes.GetClassList(ctx)
	if err != nil {
		return nil, err
	}

	return &classListView{
		elementMapView: newElementMapView(
			snapshot,
			func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
				enabled, err := runtime.CastBoolean(value)
				if err != nil {
					return runtime.None, false, err
				}

				name := runtime.ToString(key)
				if err := classes.SetClass(ctx, name, enabled); err != nil {
					return runtime.None, false, err
				}

				if !enabled {
					return runtime.None, true, nil
				}

				return runtime.True, false, nil
			},
			func(ctx context.Context, key runtime.Value) error {
				return classes.SetClass(ctx, runtime.ToString(key), runtime.False)
			},
		),
	}, nil
}
