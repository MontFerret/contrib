package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	if err := el.ensureAttached(); err != nil {
		return nil, err
	}

	return subscribeDOMTargetEvents(
		ctx,
		el.client.Runtime,
		el.eval,
		el.id,
		subscription,
	)
}
