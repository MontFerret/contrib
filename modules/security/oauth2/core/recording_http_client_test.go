package core

import (
	"context"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

type recordingTokenHTTPClient struct {
	do func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error)
}

func newRecordingTokenHTTPClient(
	do func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error),
) *recordingTokenHTTPClient {
	return &recordingTokenHTTPClient{do: do}
}

func (c *recordingTokenHTTPClient) Do(
	ctx context.Context,
	request *ferrethttp.Request,
) (*ferrethttp.Response, error) {
	return c.do(ctx, request)
}
