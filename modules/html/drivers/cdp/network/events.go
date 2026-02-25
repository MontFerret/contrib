package network

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/html/drivers/cdp/events"
)

var (
	responseReceivedEvent = events.New("response_received")
	requestPausedEvent    = events.New("request_paused")
)

func createResponseReceivedStreamFactory(client *cdp.Client) events.SourceFactory {
	return events.NewStreamSourceFactory(responseReceivedEvent, func(ctx context.Context) (rpcc.Stream, error) {
		return client.Network.ResponseReceived(ctx)
	}, func(stream rpcc.Stream) (any, error) {
		return stream.(network.ResponseReceivedClient).Recv()
	})
}

func createRequestPausedStreamFactory(client *cdp.Client) events.SourceFactory {
	return events.NewStreamSourceFactory(requestPausedEvent, func(ctx context.Context) (rpcc.Stream, error) {
		return client.Fetch.RequestPaused(ctx)
	}, func(stream rpcc.Stream) (any, error) {
		return stream.(fetch.RequestPausedClient).Recv()
	})
}
