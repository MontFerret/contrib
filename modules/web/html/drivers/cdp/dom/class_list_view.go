package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type classListView struct {
	*elementMapView
}

func newClassListView(ctx context.Context, el *HTMLElement) (*classListView, error) {
	snapshotValue, err := el.eval.EvalValue(ctx, templates.GetClassList(el.id))
	if err != nil {
		return nil, err
	}

	snapshot, err := runtime.ToMap(ctx, snapshotValue)
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
				if err := el.eval.Eval(ctx, templates.SetClass(el.id, name, enabled)); err != nil {
					return runtime.None, false, err
				}

				if !enabled {
					return runtime.None, true, nil
				}

				return runtime.True, false, nil
			},
			func(ctx context.Context, key runtime.Value) error {
				return el.eval.Eval(ctx, templates.SetClass(el.id, runtime.ToString(key), runtime.False))
			},
		),
	}, nil
}
