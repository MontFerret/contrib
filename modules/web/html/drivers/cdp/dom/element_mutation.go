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

func (el *HTMLElement) GetTextContent(ctx context.Context) (runtime.String, error) {
	out, err := el.eval.EvalValue(ctx, templates.GetTextContent(el.id))
	if err != nil {
		return runtime.EmptyString, err
	}

	return runtime.ToString(out), nil
}

func (el *HTMLElement) SetTextContent(ctx context.Context, textContent runtime.String) error {
	return el.eval.Eval(
		ctx,
		templates.SetTextContent(el.id, textContent),
	)
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
