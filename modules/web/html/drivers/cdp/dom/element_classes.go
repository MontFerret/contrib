package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementClasses struct {
	eval elementEvaluator
	id   cdpruntime.RemoteObjectID
}

func newElementClasses(exec elementEvaluator, id cdpruntime.RemoteObjectID) *elementClasses {
	return &elementClasses{
		eval: exec,
		id:   id,
	}
}

func (classes *elementClasses) GetClassList(ctx context.Context) (runtime.Map, error) {
	out, err := classes.eval.EvalValue(ctx, templates.GetClassList(classes.id))
	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (classes *elementClasses) SetClass(ctx context.Context, name runtime.String, enabled runtime.Boolean) error {
	return classes.eval.Eval(ctx, templates.SetClass(classes.id, name, enabled))
}

func (classes *elementClasses) SetClasses(ctx context.Context, values runtime.List) error {
	current, err := classes.GetClassList(ctx)
	if err != nil {
		return err
	}

	if err := current.ForEach(ctx, func(ctx context.Context, _ runtime.Value, key runtime.Value) (runtime.Boolean, error) {
		if err := classes.SetClass(ctx, runtime.ToString(key), runtime.False); err != nil {
			return runtime.False, err
		}

		return runtime.True, nil
	}); err != nil {
		return err
	}

	return values.ForEach(ctx, func(ctx context.Context, value runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
		if err := classes.SetClass(ctx, runtime.ToString(value), runtime.True); err != nil {
			return runtime.False, err
		}

		return runtime.True, nil
	})
}
