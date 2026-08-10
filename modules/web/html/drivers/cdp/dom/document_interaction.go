package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) MoveMouseByXY(ctx context.Context, x, y runtime.Float) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		value, err := state.input.MoveMouseByXY(ctx, x, y)
		return value, err
	})
}

func (doc *HTMLDocument) ScrollTop(ctx context.Context, options drivers.ScrollOptions) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		return state.input.ScrollTop(ctx, options)
	})
}

func (doc *HTMLDocument) ScrollBottom(ctx context.Context, options drivers.ScrollOptions) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		return state.input.ScrollBottom(ctx, options)
	})
}

func (doc *HTMLDocument) ScrollBySelector(ctx context.Context, selector drivers.QuerySelector, options drivers.ScrollOptions) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		return state.input.ScrollIntoViewBySelector(ctx, state.element.id, selector, options)
	})
}

func (doc *HTMLDocument) Scroll(ctx context.Context, options drivers.ScrollOptions) (runtime.Boolean, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Boolean, error) {
		return state.input.ScrollByXY(ctx, options)
	})
}
