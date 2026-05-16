package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type datasetView struct {
	*elementMapView
}

func newDatasetView(ctx context.Context, el *HTMLElement) (*datasetView, error) {
	snapshotValue, err := el.eval.EvalValue(ctx, templates.GetDataset(el.id))
	if err != nil {
		return nil, err
	}

	snapshot, err := runtime.ToMap(ctx, snapshotValue)
	if err != nil {
		return nil, err
	}

	return &datasetView{
		elementMapView: newElementMapView(
			snapshot,
			func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
				name := datasetPropertyName(key)
				if value == runtime.None {
					return runtime.None, true, el.eval.Eval(ctx, templates.RemoveDatasetProperty(el.id, name))
				}

				next := runtime.ToString(value)

				return next, false, el.eval.Eval(ctx, templates.SetDatasetProperty(el.id, name, next))
			},
			func(ctx context.Context, key runtime.Value) error {
				return el.eval.Eval(ctx, templates.RemoveDatasetProperty(el.id, datasetPropertyName(key)))
			},
		),
	}, nil
}
