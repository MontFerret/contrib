package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementDataset struct {
	eval elementEvaluator
	id   cdpruntime.RemoteObjectID
}

func newElementDataset(exec elementEvaluator, id cdpruntime.RemoteObjectID) *elementDataset {
	return &elementDataset{
		eval: exec,
		id:   id,
	}
}

func (dataset *elementDataset) GetDataset(ctx context.Context) (runtime.Map, error) {
	out, err := dataset.eval.EvalValue(ctx, templates.GetDataset(dataset.id))
	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (dataset *elementDataset) SetDatasetProperty(ctx context.Context, name, value runtime.String) error {
	return dataset.eval.Eval(ctx, templates.SetDatasetProperty(dataset.id, name, value))
}

func (dataset *elementDataset) RemoveDatasetProperty(ctx context.Context, name runtime.String) error {
	return dataset.eval.Eval(ctx, templates.RemoveDatasetProperty(dataset.id, name))
}

func (dataset *elementDataset) SetDataset(ctx context.Context, values runtime.Map) error {
	current, err := dataset.GetDataset(ctx)
	if err != nil {
		return err
	}

	if err := current.ForEach(ctx, func(ctx context.Context, _ runtime.Value, key runtime.Value) (runtime.Boolean, error) {
		if err := dataset.RemoveDatasetProperty(ctx, datasetPropertyName(key)); err != nil {
			return runtime.False, err
		}

		return runtime.True, nil
	}); err != nil {
		return err
	}

	return values.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		if err := dataset.SetDatasetProperty(ctx, datasetPropertyName(key), runtime.ToString(value)); err != nil {
			return runtime.False, err
		}

		return runtime.True, nil
	})
}
