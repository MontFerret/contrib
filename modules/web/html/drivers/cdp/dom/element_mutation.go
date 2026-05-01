package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) SetValue(ctx context.Context, value runtime.Value) error {
	return el.eval.Eval(ctx, templates.SetValue(el.id, value))
}

func (el *HTMLElement) GetStyles(ctx context.Context) (runtime.Map, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetStyles(el.id))
	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (el *HTMLElement) GetStyle(ctx context.Context, name runtime.String) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetStyle(el.id, name))
}

func (el *HTMLElement) SetStyles(ctx context.Context, styles runtime.Map) error {
	return el.eval.Eval(ctx, templates.SetStyles(el.id, styles))
}

func (el *HTMLElement) SetStyle(ctx context.Context, name, value runtime.String) error {
	return el.eval.Eval(ctx, templates.SetStyle(el.id, name, value))
}

func (el *HTMLElement) RemoveStyle(ctx context.Context, names ...runtime.String) error {
	return el.eval.Eval(ctx, templates.RemoveStyles(el.id, names))
}

func (el *HTMLElement) GetAttributes(ctx context.Context) (runtime.Map, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetAttributes(el.id))
	if err != nil {
		return runtime.NewObject(), err
	}

	return runtime.ToMap(ctx, out)
}

func (el *HTMLElement) GetAttribute(ctx context.Context, name runtime.String) (runtime.Value, error) {
	return el.eval.EvalValue(ctx, templates.GetAttribute(el.id, name))
}

func (el *HTMLElement) SetAttributes(ctx context.Context, attrs runtime.Map) error {
	return el.eval.Eval(ctx, templates.SetAttributes(el.id, attrs))
}

func (el *HTMLElement) SetAttribute(ctx context.Context, name, value runtime.String) error {
	return el.eval.Eval(ctx, templates.SetAttribute(el.id, name, value))
}

func (el *HTMLElement) RemoveAttribute(ctx context.Context, names ...runtime.String) error {
	return el.eval.Eval(ctx, templates.RemoveAttributes(el.id, names))
}

func (el *HTMLElement) GetInnerText(ctx context.Context) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerText(el.id))
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerText(ctx context.Context, innerText runtime.String) error {
	return el.eval.Eval(
		ctx,
		templates.SetInnerText(el.id, innerText),
	)
}

func (el *HTMLElement) GetInnerTextBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerTextBySelector(el.id, selector))
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerTextBySelector(ctx context.Context, selector drivers.QuerySelector, innerText runtime.String) error {
	return el.eval.Eval(
		ctx,
		templates.SetInnerTextBySelector(el.id, selector, innerText),
	)
}

func (el *HTMLElement) GetInnerTextBySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerTextBySelectorAll(el.id, selector))
	if err != nil {
		return runtime.EmptyArray(), err
	}

	return runtime.ToList(ctx, out)
}

func (el *HTMLElement) GetInnerHTML(ctx context.Context) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerHTML(el.id))
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerHTML(ctx context.Context, innerHTML runtime.String) error {
	return el.eval.Eval(ctx, templates.SetInnerHTML(el.id, innerHTML))
}

func (el *HTMLElement) GetInnerHTMLBySelector(ctx context.Context, selector drivers.QuerySelector) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerHTMLBySelector(el.id, selector))
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetInnerHTMLBySelector(ctx context.Context, selector drivers.QuerySelector, innerHTML runtime.String) error {
	return el.eval.Eval(ctx, templates.SetInnerHTMLBySelector(el.id, selector, innerHTML))
}

func (el *HTMLElement) GetInnerHTMLBySelectorAll(ctx context.Context, selector drivers.QuerySelector) (runtime.List, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetInnerHTMLBySelectorAll(el.id, selector))
	if err != nil {
		return runtime.EmptyArray(), err
	}

	return runtime.ToList(ctx, out)
}
