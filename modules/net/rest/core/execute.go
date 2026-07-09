package core

import (
	"context"

	ferretnet "github.com/MontFerret/ferret/v2/pkg/net"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func executeQuery(ctx context.Context, client *Client, q runtime.Query) (runtime.Value, bool, error) {
	if err := validateQueryDialect(q.Kind); err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	requestData, err := DecodeRequestData(ctx, q.Params)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	options, err := DecodeExecutionOptions(ctx, client.config, q.Options)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	requestURL, err := resolveRequestURL(ctx, client.config.BaseURL, q.Expression.String(), requestData.Query)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	body, contentType, err := encodeRequestBody(ctx, requestData.Body, options.RequestEncoding)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	httpClient, err := ferretnet.HTTPClientFrom(ctx)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	requestCtx := ctx
	cancel := func() {}

	if options.Timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, options.Timeout)
	}

	defer cancel()

	headers := mergeHeaders(client.config.Headers, requestData.Headers)
	if contentType != "" && !hasHeader(headers, "Content-Type") {
		headers.Set("Content-Type", contentType)
	}

	resp, err := httpClient.Do(requestCtx, &ferrethttp.Request{
		Method:  requestData.Method,
		URL:     requestURL,
		Headers: ferrethttp.Headers(headers),
		Body:    body,
	})
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	value, flatten, err := decodeHTTPResponse(ctx, requestURL, resp, options)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	return value, flatten, nil
}
