package openai

import (
	"context"
	"net/http"
	"sync"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type fakeHTTPClient struct {
	handler  func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error)
	requests []*ferrethttp.Request
	mu       sync.Mutex
}

func newFakeHTTPClient(
	handler func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error),
) *fakeHTTPClient {
	return &fakeHTTPClient{handler: handler}
}

func (f *fakeHTTPClient) Do(ctx context.Context, request *ferrethttp.Request) (*ferrethttp.Response, error) {
	copy := &ferrethttp.Request{
		Method:  request.Method,
		URL:     request.URL,
		Headers: ferrethttp.Headers(http.Header(request.Headers).Clone()),
		Body:    append([]byte(nil), request.Body...),
	}

	f.mu.Lock()
	f.requests = append(f.requests, copy)
	f.mu.Unlock()

	return f.handler(ctx, copy)
}

func (f *fakeHTTPClient) Requests() []*ferrethttp.Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]*ferrethttp.Request(nil), f.requests...)
}
