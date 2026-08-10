package dom

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (el *HTMLElement) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	var stream runtime.Stream

	err := el.executor.run(ctx, func() error {
		var err error
		stream, err = subscribeDOMTargetEvents(
			ctx,
			el.client.Runtime,
			el.executor,
			el.id,
			subscription,
		)

		return err
	})

	return stream, err
}
