package cdp

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpnet "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (p *HTMLPage) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	switch subscription.EventName {
	case drivers.NavigationEvent:
		p.mu.Lock()
		defer p.mu.Unlock()

		stream, err := p.navigationStream(ctx)
		if err != nil {
			return nil, err
		}

		opts, err := p.parseNavigationSubscriptionOptions(ctx, subscription.Options)
		if err != nil {
			_ = stream.Close()
			return nil, err
		}

		if opts.FrameID == "" && opts.URL == nil {
			return stream, nil
		}

		return newFilteredNavigationEventStream(stream, func(evt *cdpnet.NavigationEvent) bool {
			return matchNavigationEvent(evt, opts)
		}), nil
	case drivers.RequestEvent:
		return p.network.OnRequest(ctx)
	case drivers.ResponseEvent:
		return p.network.OnResponse(ctx)
	default:
		return nil, runtime.Errorf(runtime.ErrInvalidOperation, "unknown event name: %s", subscription.EventName)
	}
}

func (p *HTMLPage) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return runtime.Error(runtime.ErrNotImplemented, "HTMLPage.Dispatch")
}
