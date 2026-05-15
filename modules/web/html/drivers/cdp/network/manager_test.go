package network_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/fetch"
	network2 "github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
)

type (
	PageAPI struct {
		cdp.Page
		frameNavigated func(ctx context.Context) (page.FrameNavigatedClient, error)
		mock.Mock
	}

	NetworkAPI struct {
		cdp.Network
		requestWillBeSent      func(ctx context.Context) (network2.RequestWillBeSentClient, error)
		responseReceived       func(ctx context.Context) (network2.ResponseReceivedClient, error)
		loadingFinished        func(ctx context.Context) (network2.LoadingFinishedClient, error)
		loadingFailed          func(ctx context.Context) (network2.LoadingFailedClient, error)
		requestServedFromCache func(ctx context.Context) (network2.RequestServedFromCacheClient, error)
		getResponseBody        func(ctx context.Context, args *network2.GetResponseBodyArgs) (*network2.GetResponseBodyReply, error)
		getCookies             func(ctx context.Context, args *network2.GetCookiesArgs) (*network2.GetCookiesReply, error)
		setExtraHTTPHeaders    func(ctx context.Context, args *network2.SetExtraHTTPHeadersArgs) error
		mock.Mock
	}

	FetchAPI struct {
		cdp.Fetch
		enable        func(context.Context, *fetch.EnableArgs) error
		disable       func(context.Context) error
		requestPaused func(context.Context) (fetch.RequestPausedClient, error)
		mock.Mock
	}

	TestEventStream struct {
		closeErr error
		ready    chan struct{}
		message  chan any
		mock.Mock
		closeOnce sync.Once
	}

	FrameNavigatedClient struct {
		*TestEventStream
	}

	ResponseReceivedClient struct {
		*TestEventStream
	}

	RequestWillBeSentClient struct {
		*TestEventStream
	}

	LoadingFinishedClient struct {
		*TestEventStream
	}

	LoadingFailedClient struct {
		*TestEventStream
	}

	RequestServedFromCacheClient struct {
		*TestEventStream
	}

	RequestPausedClient struct {
		*TestEventStream
	}
)

func (api *PageAPI) FrameNavigated(ctx context.Context) (page.FrameNavigatedClient, error) {
	return api.frameNavigated(ctx)
}

func (api *NetworkAPI) ResponseReceived(ctx context.Context) (network2.ResponseReceivedClient, error) {
	if api.responseReceived == nil {
		return NewResponseReceivedClient(), nil
	}

	return api.responseReceived(ctx)
}

func (api *NetworkAPI) RequestWillBeSent(ctx context.Context) (network2.RequestWillBeSentClient, error) {
	if api.requestWillBeSent == nil {
		return NewRequestWillBeSentClient(), nil
	}

	return api.requestWillBeSent(ctx)
}

func (api *NetworkAPI) LoadingFinished(ctx context.Context) (network2.LoadingFinishedClient, error) {
	if api.loadingFinished == nil {
		return NewLoadingFinishedClient(), nil
	}

	return api.loadingFinished(ctx)
}

func (api *NetworkAPI) LoadingFailed(ctx context.Context) (network2.LoadingFailedClient, error) {
	if api.loadingFailed == nil {
		return NewLoadingFailedClient(), nil
	}

	return api.loadingFailed(ctx)
}

func (api *NetworkAPI) RequestServedFromCache(ctx context.Context) (network2.RequestServedFromCacheClient, error) {
	if api.requestServedFromCache == nil {
		return NewRequestServedFromCacheClient(), nil
	}

	return api.requestServedFromCache(ctx)
}

func (api *NetworkAPI) GetResponseBody(
	ctx context.Context,
	args *network2.GetResponseBodyArgs,
) (*network2.GetResponseBodyReply, error) {
	if api.getResponseBody == nil {
		return nil, errors.New("response body is not available")
	}

	return api.getResponseBody(ctx, args)
}

func (api *NetworkAPI) GetCookies(ctx context.Context, args *network2.GetCookiesArgs) (*network2.GetCookiesReply, error) {
	return api.getCookies(ctx, args)
}

func (api *NetworkAPI) SetExtraHTTPHeaders(ctx context.Context, args *network2.SetExtraHTTPHeadersArgs) error {
	return api.setExtraHTTPHeaders(ctx, args)
}

func (api *FetchAPI) Enable(ctx context.Context, args *fetch.EnableArgs) error {
	if api.enable == nil {
		return nil
	}

	return api.enable(ctx, args)
}

func (api *FetchAPI) Disable(ctx context.Context) error {
	if api.disable == nil {
		return nil
	}

	return api.disable(ctx)
}

func (api *FetchAPI) RequestPaused(ctx context.Context) (fetch.RequestPausedClient, error) {
	return api.requestPaused(ctx)
}

func NewTestEventStream() *TestEventStream {
	return NewBufferedTestEventStream(0)
}

func NewBufferedTestEventStream(buffer int) *TestEventStream {
	es := new(TestEventStream)
	es.ready = make(chan struct{}, buffer)
	es.message = make(chan any, buffer)
	return es
}

func (stream *TestEventStream) Ready() <-chan struct{} {
	return stream.ready
}

func (stream *TestEventStream) RecvMsg(i any) error {
	return nil
}

func (stream *TestEventStream) Message() (any, error) {
	msg, ok := <-stream.message
	if !ok {
		return nil, io.EOF
	}

	return msg, nil
}

func (stream *TestEventStream) Close() error {
	stream.closeOnce.Do(func() {
		if len(stream.ExpectedCalls) > 0 {
			args := stream.MethodCalled("Close")
			if len(args) > 0 {
				stream.closeErr = args.Error(0)
			}
		}

		close(stream.message)
		close(stream.ready)
	})

	return stream.closeErr
}

func (stream *TestEventStream) Emit(msg any) {
	stream.ready <- struct{}{}
	stream.message <- msg
}

func NewFrameNavigatedClient() *FrameNavigatedClient {
	return &FrameNavigatedClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *FrameNavigatedClient) Recv() (*page.FrameNavigatedReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*page.FrameNavigatedReply)

	if !ok {
		return nil, fmt.Errorf("invalid frame navigated message type %T", msg)
	}

	return repl, nil
}

func NewResponseReceivedClient() *ResponseReceivedClient {
	return &ResponseReceivedClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *ResponseReceivedClient) Recv() (*network2.ResponseReceivedReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*network2.ResponseReceivedReply)

	if !ok {
		return nil, fmt.Errorf("invalid response received message type %T", msg)
	}

	return repl, nil
}

func NewRequestWillBeSentClient() *RequestWillBeSentClient {
	return &RequestWillBeSentClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *RequestWillBeSentClient) Recv() (*network2.RequestWillBeSentReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*network2.RequestWillBeSentReply)

	if !ok {
		return nil, fmt.Errorf("invalid request will be sent message type %T", msg)
	}

	return repl, nil
}

func NewLoadingFinishedClient() *LoadingFinishedClient {
	return &LoadingFinishedClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *LoadingFinishedClient) Recv() (*network2.LoadingFinishedReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*network2.LoadingFinishedReply)

	if !ok {
		return nil, fmt.Errorf("invalid loading finished message type %T", msg)
	}

	return repl, nil
}

func NewLoadingFailedClient() *LoadingFailedClient {
	return &LoadingFailedClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *LoadingFailedClient) Recv() (*network2.LoadingFailedReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*network2.LoadingFailedReply)

	if !ok {
		return nil, fmt.Errorf("invalid loading failed message type %T", msg)
	}

	return repl, nil
}

func NewRequestServedFromCacheClient() *RequestServedFromCacheClient {
	return &RequestServedFromCacheClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *RequestServedFromCacheClient) Recv() (*network2.RequestServedFromCacheReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*network2.RequestServedFromCacheReply)

	if !ok {
		return nil, fmt.Errorf("invalid request served from cache message type %T", msg)
	}

	return repl, nil
}

func NewRequestPausedClient() *RequestPausedClient {
	return &RequestPausedClient{
		TestEventStream: NewTestEventStream(),
	}
}

func (stream *RequestPausedClient) Recv() (*fetch.RequestPausedReply, error) {
	msg, err := stream.Message()
	if err != nil {
		return nil, err
	}

	repl, ok := msg.(*fetch.RequestPausedReply)

	if !ok {
		return nil, fmt.Errorf("invalid request paused message type %T", msg)
	}

	return repl, nil
}

func TestManager(t *testing.T) {
	Convey("Network manager", t, func() {

		Convey("New", func() {
			Convey("Should close all resources on Close", func() {
				responseReceivedClient := NewResponseReceivedClient()
				responseReceivedClient.On("Close").Once().Return(nil)
				networkAPI := new(NetworkAPI)
				networkAPI.responseReceived = func(ctx context.Context) (network2.ResponseReceivedClient, error) {
					return responseReceivedClient, nil
				}
				networkAPI.setExtraHTTPHeaders = func(ctx context.Context, args *network2.SetExtraHTTPHeadersArgs) error {
					return nil
				}

				requestPausedClient := NewRequestPausedClient()
				requestPausedClient.On("Close").Once().Return(nil)
				fetchAPI := new(FetchAPI)
				fetchAPI.enable = func(ctx context.Context, args *fetch.EnableArgs) error {
					return nil
				}
				fetchAPI.requestPaused = func(ctx context.Context) (fetch.RequestPausedClient, error) {
					return requestPausedClient, nil
				}

				client := &cdp.Client{
					Network: networkAPI,
					Fetch:   fetchAPI,
				}

				mgr, err := network.New(
					zerolog.New(os.Stdout).Level(zerolog.Disabled),
					client,
					nil,
					network.Options{
						Headers: drivers.NewHTTPHeadersWith(map[string][]string{"x-correlation-id": {"foo"}}),
						Filter: &network.Filter{
							Patterns: []drivers.ResourceFilter{
								{
									URL:  "http://google.com",
									Type: "img",
								},
							},
						},
					},
				)

				So(err, ShouldBeNil)
				So(mgr.Close(), ShouldBeNil)

				time.Sleep(time.Duration(100) * time.Millisecond)

				responseReceivedClient.AssertExpectations(t)
				requestPausedClient.AssertExpectations(t)
			})
		})

		Convey("GetCookies", func() {
			Convey("Should read cookies via the Network domain for the current page URL", func() {
				responseReceivedClient := NewResponseReceivedClient()
				responseReceivedClient.On("Close").Maybe().Return(nil)

				networkAPI := new(NetworkAPI)
				networkAPI.responseReceived = func(ctx context.Context) (network2.ResponseReceivedClient, error) {
					return responseReceivedClient, nil
				}
				networkAPI.getCookies = func(ctx context.Context, args *network2.GetCookiesArgs) (*network2.GetCookiesReply, error) {
					So(args, ShouldNotBeNil)
					So(args.URLs, ShouldResemble, []string{"http://example.com/app"})

					return &network2.GetCookiesReply{
						Cookies: []network2.Cookie{
							{
								Name:   "Session",
								Value:  "abc123",
								Path:   "/",
								Domain: "example.com",
							},
						},
					}, nil
				}

				client := &cdp.Client{
					Network: networkAPI,
				}

				mgr, err := network.New(
					zerolog.New(os.Stdout).Level(zerolog.Disabled),
					client,
					nil,
					network.Options{},
				)
				So(err, ShouldBeNil)

				cookies, err := mgr.GetCookies(context.Background(), "example.com/app")
				So(err, ShouldBeNil)
				So(cookies, ShouldNotBeNil)
				So(len(cookies.Data), ShouldEqual, 1)
				So(cookies.Data["Session"].Value, ShouldEqual, "abc123")

				So(mgr.Close(), ShouldBeNil)
				time.Sleep(time.Duration(100) * time.Millisecond)
				responseReceivedClient.AssertExpectations(t)
			})
		})
	})
}
