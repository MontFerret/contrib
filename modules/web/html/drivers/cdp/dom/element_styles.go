package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementStyles struct {
	eval elementEvaluator
	id   cdpruntime.RemoteObjectID
}

func newElementStyles(exec elementEvaluator, id cdpruntime.RemoteObjectID) *elementStyles {
	return &elementStyles{
		eval: exec,
		id:   id,
	}
}

func (styles *elementStyles) GetStyles(ctx context.Context) (runtime.Map, error) {
	out, err := styles.eval.EvalValue(ctx, templates.GetStyles(styles.id))
	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (styles *elementStyles) GetStyle(ctx context.Context, name runtime.String) (runtime.Value, error) {
	return styles.eval.EvalValue(ctx, templates.GetStyle(styles.id, name))
}

func (styles *elementStyles) SetStyles(ctx context.Context, values runtime.Map) error {
	return styles.eval.Eval(ctx, templates.SetStyles(styles.id, values))
}

func (styles *elementStyles) SetStyle(ctx context.Context, name, value runtime.String) error {
	return styles.eval.Eval(ctx, templates.SetStyle(styles.id, name, value))
}

func (styles *elementStyles) RemoveStyle(ctx context.Context, names ...runtime.String) error {
	return styles.eval.Eval(ctx, templates.RemoveStyles(styles.id, names))
}
