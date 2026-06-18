package core

import (
	"context"
	"io"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func executeQuery(ctx context.Context, client *Client, q runtime.Query) (runtime.Value, bool, error) {
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

	requestCtx := ctx
	cancel := func() {}

	if options.Timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, options.Timeout)
	}

	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, requestData.Method, requestURL, body)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	req.Header = mergeHeaders(client.config.Headers, requestData.Headers)
	if contentType != "" && !hasHeader(req.Header, "Content-Type") {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	value, flatten, err := decodeHTTPResponse(ctx, resp, responseBody, options)
	if err != nil {
		return runtime.None, false, OperationError("QUERY", err)
	}

	return value, flatten, nil
}
