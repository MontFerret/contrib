package openai

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type ferretHTTPClient struct{}

func (ferretHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("openai HTTP request is nil")
	}
	if req.Body != nil {
		defer req.Body.Close()
	}

	var body []byte
	var err error
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	client, err := ferretnet.HTTPClientFrom(req.Context())
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req.Context(), &ferrethttp.Request{
		Method:  req.Method,
		URL:     req.URL.String(),
		Headers: ferrethttp.Headers(req.Header.Clone()),
		Body:    body,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("ferret HTTP client returned a nil response")
	}

	return &http.Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Header:        http.Header(resp.Headers).Clone(),
		Body:          io.NopCloser(bytes.NewReader(resp.Body)),
		ContentLength: int64(len(resp.Body)),
		Request:       req,
	}, nil
}
