package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (doc *HTMLDocument) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	return withDocumentResult(ctx, doc, func(state *documentState) (runtime.Stream, error) {
		return subscribeDOMTargetEvents(
			ctx,
			state.client.Runtime,
			state.eval,
			state.element.id,
			subscription,
		)
	})
}
