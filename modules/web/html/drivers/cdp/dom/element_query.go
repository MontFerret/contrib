package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/queryutil"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) GetChildNodes(ctx context.Context) (runtime.List, error) {
	return el.eval.EvalElements(ctx, templates.GetChildren(el.id))
}

func (el *HTMLElement) GetChildNode(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetChildByIndex(el.id, idx))
}

func (el *HTMLElement) GetParentElement(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetParent(el.id))
}

func (el *HTMLElement) GetPreviousElementSibling(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetPreviousElementSibling(el.id))
}

func (el *HTMLElement) GetNextElementSibling(ctx context.Context) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.GetNextElementSibling(el.id))
}

func (el *HTMLElement) QuerySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Value, error) {
	return el.eval.EvalElement(ctx, templates.QuerySelector(el.id, selector))
}

func (el *HTMLElement) QuerySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	return el.eval.EvalElements(ctx, templates.QuerySelectorAll(el.id, selector))
}

func (el *HTMLElement) XPath(ctx context.Context, expression runtime.String) (result runtime.Value, err error) {
	return el.eval.EvalValue(ctx, templates.XPath(el.id, expression))
}

func (el *HTMLElement) CountBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Int, error) {
	out, err := el.eval.EvalValue(ctx, templates.CountBySelector(el.id, selector))
	if err != nil {
		return runtime.ZeroInt, err
	}

	return runtime.ToInt(ctx, out)
}

func (el *HTMLElement) ExistsBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.Boolean, error) {
	out, err := el.eval.EvalValue(ctx, templates.ExistsBySelector(el.id, selector))
	if err != nil {
		return runtime.False, err
	}

	return runtime.ToBoolean(out), nil
}

func (el *HTMLElement) Query(ctx context.Context, q runtime.Query) (runtime.List, error) {
	switch queryutil.Parse(string(q.Kind)) {
	case queryutil.CSS:
		fn, err := templates.CSSX(el.id, q.Payload)
		if err != nil {
			return runtime.NewArray(0), err
		}

		val, err := el.eval.EvalValue(ctx, fn)
		if err != nil {
			return runtime.NewArray(0), err
		}

		return runtime.ToList(ctx, val)
	case queryutil.XPath:
		out, err := el.XPath(ctx, q.Payload)
		if err != nil {
			return runtime.NewArray(0), err
		}

		list, ok := out.(runtime.List)
		if ok {
			return list, nil
		}

		return runtime.NewArrayWith(out), nil
	default:
		return nil, runtime.Error(runtime.ErrInvalidArgument, "unsupported query kind")
	}
}
