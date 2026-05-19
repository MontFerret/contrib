package dom

import (
	"context"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type elementAttributes struct {
	eval elementEvaluator
	id   cdpruntime.RemoteObjectID
}

func newElementAttributes(exec elementEvaluator, id cdpruntime.RemoteObjectID) *elementAttributes {
	return &elementAttributes{
		eval: exec,
		id:   id,
	}
}

func (attrs *elementAttributes) GetAttributes(ctx context.Context) (runtime.Map, error) {
	out, err := attrs.eval.EvalValue(ctx, templates.GetAttributes(attrs.id))
	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (attrs *elementAttributes) GetAttribute(ctx context.Context, name runtime.String) (runtime.Value, error) {
	return attrs.eval.EvalValue(ctx, templates.GetAttribute(attrs.id, name))
}

func (attrs *elementAttributes) SetAttributes(ctx context.Context, values runtime.Map) error {
	return attrs.eval.Eval(ctx, templates.SetAttributes(attrs.id, values))
}

func (attrs *elementAttributes) SetAttribute(ctx context.Context, name, value runtime.String) error {
	return attrs.eval.Eval(ctx, templates.SetAttribute(attrs.id, name, value))
}

func (attrs *elementAttributes) RemoveAttribute(ctx context.Context, names ...runtime.String) error {
	return attrs.eval.Eval(ctx, templates.RemoveAttributes(attrs.id, names))
}
