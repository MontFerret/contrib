package cdp

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	eventName := subscription.EventName.String()
	if !drivers.IsObservableEvent(eventName) {
		return nil, invalidObservableEventNameError(eventName)
	}

	switch eventName {
	case drivers.NavigationEvent:
		p.mu.Lock()
		defer p.mu.Unlock()

		opts, err := p.parseNavigationSubscriptionOptions(ctx, subscription.Options)
		if err != nil {
			return nil, err
		}

		return p.navigationStream(ctx, opts)
	case drivers.RequestEvent:
		return p.network.OnRequest(ctx)
	case drivers.ResponseEvent:
		return p.network.OnResponse(ctx)
	default:
		return p.network.OnEvent(ctx, subscription.EventName, subscription.Options)
	}
}

func (p *HTMLPage) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return p.getCurrentDocument().Dispatch(ctx, event)
}
