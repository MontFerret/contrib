package network

import (
	"context"

	"github.com/rs/zerolog"

	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func newNavigationEventStream(logger zerolog.Logger, sessions *cdpsession.Manager) runtime.Stream {
	return newSessionRuntimeStream(logger, sessions, func(ctx context.Context, client *cdpsession.Client) (runtime.Stream, error) {
		frameStream, err := client.CDP.Page.FrameNavigated(ctx)
		if err != nil {
			return nil, err
		}

		docStream, err := client.CDP.Page.NavigatedWithinDocument(ctx)
		if err != nil {
			_ = frameStream.Close()
			return nil, err
		}

		return newSessionNavigationEventStream(logger, client.CDP, frameStream, docStream), nil
	})
}

func newRequestEventStream(logger zerolog.Logger, sessions *cdpsession.Manager) runtime.Stream {
	return newSessionRuntimeStream(logger, sessions, func(ctx context.Context, client *cdpsession.Client) (runtime.Stream, error) {
		stream, err := client.CDP.Network.RequestWillBeSent(ctx)
		if err != nil {
			return nil, err
		}

		return newRequestWillBeSentStream(logger, stream), nil
	})
}

func newResponseEventStream(logger zerolog.Logger, sessions *cdpsession.Manager) runtime.Stream {
	return newSessionRuntimeStream(logger, sessions, func(ctx context.Context, client *cdpsession.Client) (runtime.Stream, error) {
		stream, err := client.CDP.Network.ResponseReceived(ctx)
		if err != nil {
			return nil, err
		}

		return newResponseReceivedReader(logger, client.CDP, stream), nil
	})
}
