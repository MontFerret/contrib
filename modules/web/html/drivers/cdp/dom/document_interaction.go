package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) MoveMouseByXY(ctx context.Context, x, y runtime.Float) error {
	return doc.input.MoveMouseByXY(ctx, x, y)
}

func (doc *HTMLDocument) ScrollTop(ctx context.Context, options drivers.ScrollOptions) error {
	return doc.input.ScrollTop(ctx, options)
}

func (doc *HTMLDocument) ScrollBottom(ctx context.Context, options drivers.ScrollOptions) error {
	return doc.input.ScrollBottom(ctx, options)
}

func (doc *HTMLDocument) ScrollBySelector(ctx context.Context, selector drivers.QuerySelector, options drivers.ScrollOptions) error {
	return doc.input.ScrollIntoViewBySelector(ctx, doc.element.id, selector, options)
}

func (doc *HTMLDocument) Scroll(ctx context.Context, options drivers.ScrollOptions) error {
	return doc.input.ScrollByXY(ctx, options)
}
