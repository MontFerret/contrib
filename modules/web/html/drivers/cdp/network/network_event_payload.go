package network

import (
	"context"
	"encoding/base64"

	cdpnetwork "github.com/mafredri/cdp/protocol/network"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func buildNetworkEventPayload(
	ctx context.Context,
	logger zerolog.Logger,
	event networkEvent,
	options networkEventOptions,
) runtime.Value {
	props := map[string]runtime.Value{
		"event":             runtime.NewString(event.name),
		"requestId":         runtime.NewString(string(event.requestID)),
		"loaderId":          runtime.NewString(string(event.loaderID)),
		"frameId":           runtime.NewString(event.frameID),
		"url":               runtime.NewString(event.url),
		"method":            runtime.NewString(event.method),
		"type":              runtime.NewString(event.resourceType),
		"status":            runtime.NewInt(event.status),
		"statusText":        runtime.NewString(event.statusText),
		"mimeType":          runtime.NewString(event.mimeType),
		"headers":           headersRuntimeValue(event.headers),
		"requestHeaders":    headersRuntimeValue(event.requestHeaders),
		"failed":            runtime.NewBoolean(event.failed),
		"errorText":         runtime.NewString(event.errorText),
		"canceled":          runtime.NewBoolean(event.canceled),
		"blockedReason":     runtime.NewString(event.blockedReason),
		"fromCache":         runtime.NewBoolean(event.fromCache),
		"fromDiskCache":     runtime.NewBoolean(event.fromDiskCache),
		"fromServiceWorker": runtime.NewBoolean(event.fromServiceWorker),
		"fromPrefetchCache": runtime.NewBoolean(event.fromPrefetchCache),
		"encodedDataLength": runtime.NewFloat(event.encodedDataLength),
		"timestamp":         runtime.NewFloat(event.timestamp),
		"wallTime":          runtime.NewFloat(event.wallTime),
	}

	if event.name == drivers.NetworkRequestFinishedEvent && options.captureBody {
		body, truncated := captureNetworkEventBody(ctx, logger, event, options.bodyLimit)
		props["body"] = body
		props["bodyTruncated"] = runtime.NewBoolean(truncated)
	}

	return runtime.NewObjectWith(props)
}

func buildNetworkIdlePayload(options networkIdleOptions, inflight int) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"event":       runtime.NewString(drivers.NetworkIdleEvent),
		"inflight":    runtime.NewInt(inflight),
		"maxInflight": runtime.NewInt(options.maxInflight),
		"quiet":       runtime.NewInt64(options.quiet.Milliseconds()),
		"types":       resourceTypesRuntimeValue(options.typeList),
	})
}

func captureNetworkEventBody(
	ctx context.Context,
	logger zerolog.Logger,
	event networkEvent,
	bodyLimit int,
) (runtime.Value, bool) {
	if event.client == nil || event.client.Network == nil {
		return runtime.None, false
	}

	repl, err := event.client.Network.GetResponseBody(ctx, cdpnetwork.NewGetResponseBodyArgs(event.requestID))
	if err != nil {
		logger.Warn().
			Err(err).
			Str("request_id", string(event.requestID)).
			Str("url", event.url).
			Msg("failed to get network event response body")

		return runtime.None, false
	}

	var body []byte
	if repl.Base64Encoded {
		body, err = base64.StdEncoding.DecodeString(repl.Body)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("request_id", string(event.requestID)).
				Str("url", event.url).
				Msg("failed to decode network event response body")

			return runtime.None, false
		}
	} else {
		body = []byte(repl.Body)
	}

	truncated := false
	if len(body) > bodyLimit {
		body = body[:bodyLimit]
		truncated = true
	}

	return runtime.NewBinary(body), truncated
}

func headersRuntimeValue(headers *drivers.HTTPHeaders) runtime.Value {
	if headers == nil {
		return runtime.None
	}

	return headers
}

func resourceTypesRuntimeValue(types []string) runtime.Value {
	values := make([]runtime.Value, 0, len(types))
	for _, resourceType := range types {
		values = append(values, runtime.NewString(resourceType))
	}

	return runtime.NewArrayWith(values...)
}
