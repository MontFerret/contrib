package cdp

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/mafredri/cdp"
	cdpnetwork "github.com/mafredri/cdp/protocol/network"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpnet "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type pageEventNetworkAPI struct {
	cdp.Network
}

type pageEventNetworkStream struct {
	ready     chan struct{}
	closeOnce sync.Once
}

type pageEventRequestWillBeSentClient struct {
	*pageEventNetworkStream
}

type pageEventResponseReceivedClient struct {
	*pageEventNetworkStream
}

type pageEventLoadingFinishedClient struct {
	*pageEventNetworkStream
}

type pageEventLoadingFailedClient struct {
	*pageEventNetworkStream
}

type pageEventRequestServedFromCacheClient struct {
	*pageEventNetworkStream
}

func (api *pageEventNetworkAPI) RequestWillBeSent(context.Context) (cdpnetwork.RequestWillBeSentClient, error) {
	return &pageEventRequestWillBeSentClient{pageEventNetworkStream: newPageEventNetworkStream()}, nil
}

func (api *pageEventNetworkAPI) ResponseReceived(context.Context) (cdpnetwork.ResponseReceivedClient, error) {
	return &pageEventResponseReceivedClient{pageEventNetworkStream: newPageEventNetworkStream()}, nil
}

func (api *pageEventNetworkAPI) LoadingFinished(context.Context) (cdpnetwork.LoadingFinishedClient, error) {
	return &pageEventLoadingFinishedClient{pageEventNetworkStream: newPageEventNetworkStream()}, nil
}

func (api *pageEventNetworkAPI) LoadingFailed(context.Context) (cdpnetwork.LoadingFailedClient, error) {
	return &pageEventLoadingFailedClient{pageEventNetworkStream: newPageEventNetworkStream()}, nil
}

func (api *pageEventNetworkAPI) RequestServedFromCache(
	context.Context,
) (cdpnetwork.RequestServedFromCacheClient, error) {
	return &pageEventRequestServedFromCacheClient{pageEventNetworkStream: newPageEventNetworkStream()}, nil
}

func newPageEventNetworkStream() *pageEventNetworkStream {
	return &pageEventNetworkStream{
		ready: make(chan struct{}),
	}
}

func (stream *pageEventNetworkStream) Ready() <-chan struct{} {
	return stream.ready
}

func (stream *pageEventNetworkStream) RecvMsg(any) error {
	return nil
}

func (stream *pageEventNetworkStream) Close() error {
	stream.closeOnce.Do(func() {
		close(stream.ready)
	})

	return nil
}

func (stream *pageEventRequestWillBeSentClient) Recv() (*cdpnetwork.RequestWillBeSentReply, error) {
	return nil, io.EOF
}

func (stream *pageEventResponseReceivedClient) Recv() (*cdpnetwork.ResponseReceivedReply, error) {
	return nil, io.EOF
}

func (stream *pageEventLoadingFinishedClient) Recv() (*cdpnetwork.LoadingFinishedReply, error) {
	return nil, io.EOF
}

func (stream *pageEventLoadingFailedClient) Recv() (*cdpnetwork.LoadingFailedReply, error) {
	return nil, io.EOF
}

func (stream *pageEventRequestServedFromCacheClient) Recv() (*cdpnetwork.RequestServedFromCacheReply, error) {
	return nil, io.EOF
}

func TestHTMLPageSubscribeRoutesObservableEvents(t *testing.T) {
	ctx := context.Background()
	client := &cdp.Client{Network: &pageEventNetworkAPI{}}
	manager, err := cdpnet.New(zerolog.Nop(), client, nil, cdpnet.Options{})
	if err != nil {
		t.Fatalf("unexpected network manager error: %v", err)
	}
	defer manager.Close()

	page := NewHTMLPage(zerolog.Nop(), client, nil, manager, nil)

	for _, eventName := range drivers.SupportedObservableEvents() {
		stream, err := page.Subscribe(ctx, runtime.Subscription{
			EventName: runtime.NewString(eventName),
		})
		if err != nil {
			t.Fatalf("unexpected subscribe error for %s: %v", eventName, err)
		}

		if stream == nil {
			t.Fatalf("expected stream for %s", eventName)
		}

		if err := stream.Close(); err != nil {
			t.Fatalf("unexpected stream close error for %s: %v", eventName, err)
		}
	}
}

func TestHTMLPageSubscribeRejectsUnknownEventName(t *testing.T) {
	ctx := context.Background()
	client := &cdp.Client{Network: &pageEventNetworkAPI{}}
	manager, err := cdpnet.New(zerolog.Nop(), client, nil, cdpnet.Options{})
	if err != nil {
		t.Fatalf("unexpected network manager error: %v", err)
	}
	defer manager.Close()

	page := NewHTMLPage(zerolog.Nop(), client, nil, manager, nil)

	_, err = page.Subscribe(ctx, runtime.Subscription{EventName: runtime.NewString("network.unknown")})
	if err == nil {
		t.Fatal("expected unknown event error")
	}

	if !strings.Contains(err.Error(), "unknown event name: network.unknown") {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEvents := strings.Join(drivers.SupportedObservableEvents(), ", ")
	if !strings.Contains(err.Error(), "supported events: "+expectedEvents) {
		t.Fatalf("expected supported event list in error, got %v", err)
	}
}
