package core

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func decodeHTTPResponse(ctx context.Context, requestURL string, resp *ferrethttp.Response, opts ExecutionOptions) (runtime.Value, bool, error) {
	ok := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest
	if !ok && opts.ErrorMode == ErrorModeRaise {
		return runtime.None, false, fmt.Errorf("unexpected status %s", resp.Status)
	}

	decoded, err := decodeResponseBody(resp.Body, opts.ResponseEncoding)
	if err != nil {
		return runtime.None, false, fmt.Errorf("decode response body: %w", err)
	}

	fullResponse := opts.ResponseMode == ResponseModeFull || (!ok && opts.ErrorMode == ErrorModeResponse)
	if fullResponse {
		value, err := buildFullResponse(ctx, requestURL, resp, decoded)
		if err != nil {
			return runtime.None, false, err
		}

		return value, false, nil
	}

	return decoded, true, nil
}

func buildFullResponse(ctx context.Context, requestURL string, resp *ferrethttp.Response, body runtime.Value) (runtime.Value, error) {
	out := runtime.NewObjectWith(map[string]runtime.Value{
		"ok":     runtime.NewBoolean(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest),
		"status": runtime.NewInt(resp.StatusCode),
		"body":   body,
		"url":    runtime.NewString(requestURL),
	})

	headers, err := responseHeaders(ctx, http.Header(resp.Headers))
	if err != nil {
		return runtime.None, err
	}
	if err := out.Set(ctx, runtime.NewString("headers"), headers); err != nil {
		return runtime.None, err
	}

	return out, nil
}

func responseHeaders(ctx context.Context, headers http.Header) (runtime.Value, error) {
	out := runtime.NewObjectOf(len(headers))

	for key, values := range headers {
		normalizedKey := strings.ToLower(key)
		if len(values) == 1 {
			if err := out.Set(ctx, runtime.NewString(normalizedKey), runtime.NewString(values[0])); err != nil {
				return runtime.None, err
			}

			continue
		}

		items := runtime.NewArray(len(values))
		for _, value := range values {
			if err := items.Append(ctx, runtime.NewString(value)); err != nil {
				return runtime.None, err
			}
		}
		if err := out.Set(ctx, runtime.NewString(normalizedKey), items); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}
