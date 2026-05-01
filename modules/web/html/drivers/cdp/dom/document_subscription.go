package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	return subscribeDOMTargetEvents(
		ctx,
		doc.client.Runtime,
		doc.eval,
		doc.element.id,
		subscription,
	)
}

func (doc *HTMLDocument) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return runtime.Error(runtime.ErrNotImplemented, "HTMLDocument.Dispatch")
}
