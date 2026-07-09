package core

import (
	"context"
	"sync"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type recordingHTTPClient struct {
	err         error
	lastRequest *ferrethttp.Request
	response    *ferrethttp.Response
	mu          sync.Mutex
}

func (c *recordingHTTPClient) Do(_ context.Context, req *ferrethttp.Request) (*ferrethttp.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastRequest = c.cloneRequest(req)
	if c.err != nil {
		return nil, c.err
	}

	return c.response, nil
}

func (c *recordingHTTPClient) request() *ferrethttp.Request {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cloneRequest(c.lastRequest)
}

func (c *recordingHTTPClient) cloneRequest(req *ferrethttp.Request) *ferrethttp.Request {
	if req == nil {
		return nil
	}

	clone := *req
	if req.Headers != nil {
		clone.Headers = make(ferrethttp.Headers, len(req.Headers))
		for key, values := range req.Headers {
			clone.Headers[key] = append([]string(nil), values...)
		}
	}
	clone.Body = append([]byte(nil), req.Body...)

	return &clone
}
