package dom

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) WaitForElement(ctx context.Context, selector drivers.QuerySelector, when drivers.WaitEvent) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForElement(state.element.id, selector, when))
	})
}

func (doc *HTMLDocument) WaitForClassBySelector(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForClassBySelector(state.element.id, selector, class, when))
	})
}

func (doc *HTMLDocument) WaitForClassBySelectorAll(ctx context.Context, selector drivers.QuerySelector, class runtime.String, when drivers.WaitEvent) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForClassBySelectorAll(state.element.id, selector, class, when))
	})
}

func (doc *HTMLDocument) WaitForAttributeBySelector(
	ctx context.Context,
	selector drivers.QuerySelector,
	name,
	value runtime.String,
	when drivers.WaitEvent,
) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForAttributeBySelector(state.element.id, selector, name, value, when))
	})
}

func (doc *HTMLDocument) WaitForAttributeBySelectorAll(
	ctx context.Context,
	selector drivers.QuerySelector,
	name,
	value runtime.String,
	when drivers.WaitEvent,
) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForAttributeBySelectorAll(state.element.id, selector, name, value, when))
	})
}

func (doc *HTMLDocument) WaitForStyleBySelector(ctx context.Context, selector drivers.QuerySelector, name, value runtime.String, when drivers.WaitEvent) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForStyleBySelector(state.element.id, selector, name, value, when))
	})
}

func (doc *HTMLDocument) WaitForStyleBySelectorAll(ctx context.Context, selector drivers.QuerySelector, name, value runtime.String, when drivers.WaitEvent) error {
	return withDocumentError(ctx, doc, func(state *documentState) error {
		return runDocumentWait(ctx, state, templates.WaitForStyleBySelectorAll(state.element.id, selector, name, value, when))
	})
}
