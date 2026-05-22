package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type datasetView struct {
	*elementMapView
}

func newDatasetView(ctx context.Context, dataset *elementDataset) (*datasetView, error) {
	snapshot, err := dataset.GetDataset(ctx)
	if err != nil {
		return nil, err
	}

	return &datasetView{
		elementMapView: newElementMapView(
			snapshot,
			func(ctx context.Context, key, value runtime.Value) (runtime.Value, bool, error) {
				name := datasetPropertyName(key)
				if value == runtime.None {
					return runtime.None, true, dataset.RemoveDatasetProperty(ctx, name)
				}

				next := runtime.ToString(value)

				return next, false, dataset.SetDatasetProperty(ctx, name, next)
			},
			func(ctx context.Context, key runtime.Value) error {
				return dataset.RemoveDatasetProperty(ctx, datasetPropertyName(key))
			},
		).withKeyNormalizer(func(key runtime.Value) runtime.Value {
			return datasetPropertyName(key)
		}),
	}, nil
}
