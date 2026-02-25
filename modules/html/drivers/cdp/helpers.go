package cdp

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/common"
	"github.com/mafredri/cdp/protocol/dom"

	"github.com/MontFerret/ferret/v2/pkg/runtime"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/emulation"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
)

type (
	batchFunc = func() error

	closer func() error

	pageNavigationEventStream struct {
		stream runtime.Stream
		closer closer
	}
)

func newPageNavigationEventStream(stream runtime.Stream, closer closer) runtime.Stream {
	return &pageNavigationEventStream{stream, closer}
}

func (p *pageNavigationEventStream) Close() error {
	if err := p.stream.Close(); err != nil {
		return err
	}

	return p.closer()
}

func (p *pageNavigationEventStream) Read(ctx context.Context) <-chan runtime.Message {
	return p.stream.Read(ctx)
}

func runBatch(funcs ...batchFunc) error {
	eg := errgroup.Group{}

	for _, f := range funcs {
		eg.Go(f)
	}

	return eg.Wait()
}

func enableFeatures(ctx context.Context, client *cdp.Client, params drivers.Params) error {
	if err := client.Page.Enable(ctx); err != nil {
		return err
	}

	return runBatch(
		func() error {
			return client.Page.SetLifecycleEventsEnabled(
				ctx,
				page.NewSetLifecycleEventsEnabledArgs(true),
			)
		},

		func() error {
			return client.DOM.Enable(ctx, dom.NewEnableArgs().SetIncludeWhitespace("all"))
		},

		func() error {
			return client.Runtime.Enable(ctx)
		},

		func() error {
			ua := common.GetUserAgent(params.UserAgent)

			// do not use custom user agent
			if ua == "" {
				return nil
			}

			return client.Emulation.SetUserAgentOverride(
				ctx,
				emulation.NewSetUserAgentOverrideArgs(ua),
			)
		},

		func() error {
			return client.Network.Enable(ctx, network.NewEnableArgs())
		},

		func() error {
			return client.Page.SetBypassCSP(ctx, page.NewSetBypassCSPArgs(true))
		},

		func() error {
			if params.Viewport == nil {
				return nil
			}

			orientation := emulation.ScreenOrientation{}

			if !params.Viewport.Landscape {
				orientation.Type = "portraitPrimary"
				orientation.Angle = 0
			} else {
				orientation.Type = "landscapePrimary"
				orientation.Angle = 90
			}

			scaleFactor := params.Viewport.ScaleFactor

			if scaleFactor <= 0 {
				scaleFactor = 1
			}

			deviceArgs := emulation.NewSetDeviceMetricsOverrideArgs(
				params.Viewport.Width,
				params.Viewport.Height,
				scaleFactor,
				params.Viewport.Mobile,
			).SetScreenOrientation(orientation)

			return client.Emulation.SetDeviceMetricsOverride(
				ctx,
				deviceArgs,
			)
		},
	)
}
