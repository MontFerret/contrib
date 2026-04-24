package dom

import (
	"context"
	"strings"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func (el *HTMLElement) Click(ctx context.Context, count runtime.Int) error {
	return el.input.Click(ctx, el.id, int(count))
}

func (el *HTMLElement) ClickBySelector(ctx context.Context, selector drivers.QuerySelector, count runtime.Int) error {
	return el.input.ClickBySelector(ctx, el.id, selector, count)
}

func (el *HTMLElement) ClickBySelectorAll(ctx context.Context, selector drivers.QuerySelector, count runtime.Int) error {
	elements, err := el.QuerySelectorAll(ctx, selector)
	if err != nil {
		return err
	}

	return elements.ForEach(ctx, func(ctx context.Context, value runtime.Value, idx runtime.Int) (runtime.Boolean, error) {
		found := value.(*HTMLElement)

		if e := found.Click(ctx, count); e != nil {
			err = e
			return false, e
		}

		return true, nil
	})
}

func (el *HTMLElement) Input(ctx context.Context, value runtime.Value, delay runtime.Int) error {
	name, err := el.GetNodeName(ctx)
	if err != nil {
		return err
	}

	if strings.ToLower(string(name)) != "input" {
		return runtime.Error(runtime.ErrInvalidOperation, "element is not an <input> element.")
	}

	return el.input.Type(ctx, el.id, input.TypeParams{
		Text:  value.String(),
		Clear: false,
		Delay: time.Duration(delay) * time.Millisecond,
	})
}

func (el *HTMLElement) InputBySelector(ctx context.Context, selector drivers.QuerySelector, value runtime.Value, delay runtime.Int) error {
	return el.input.TypeBySelector(ctx, el.id, selector, input.TypeParams{
		Text:  value.String(),
		Clear: false,
		Delay: time.Duration(delay) * time.Millisecond,
	})
}

func (el *HTMLElement) Press(ctx context.Context, keys []runtime.String, count runtime.Int) error {
	return el.input.Press(ctx, sdk.UnwrapStrings(keys), int(count))
}

func (el *HTMLElement) PressBySelector(ctx context.Context, selector drivers.QuerySelector, keys []runtime.String, count runtime.Int) error {
	return el.input.PressBySelector(ctx, el.id, selector, sdk.UnwrapStrings(keys), int(count))
}

func (el *HTMLElement) Clear(ctx context.Context) error {
	return el.input.Clear(ctx, el.id)
}

func (el *HTMLElement) ClearBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.ClearBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Select(ctx context.Context, value runtime.List) (runtime.List, error) {
	return el.input.Select(ctx, el.id, value)
}

func (el *HTMLElement) SelectBySelector(ctx context.Context, selector drivers.QuerySelector, value runtime.List) (runtime.List, error) {
	return el.input.SelectBySelector(ctx, el.id, selector, value)
}

func (el *HTMLElement) ScrollIntoView(ctx context.Context, options drivers.ScrollOptions) error {
	return el.input.ScrollIntoView(ctx, el.id, options)
}

func (el *HTMLElement) Focus(ctx context.Context) error {
	return el.input.Focus(ctx, el.id)
}

func (el *HTMLElement) FocusBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.FocusBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Blur(ctx context.Context) error {
	return el.input.Blur(ctx, el.id)
}

func (el *HTMLElement) BlurBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.BlurBySelector(ctx, el.id, selector)
}

func (el *HTMLElement) Hover(ctx context.Context) error {
	return el.input.MoveMouse(ctx, el.id)
}

func (el *HTMLElement) HoverBySelector(ctx context.Context, selector drivers.QuerySelector) error {
	return el.input.MoveMouseBySelector(ctx, el.id, selector)
}
