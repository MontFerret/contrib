package network

import (
	"context"
	"encoding/base64"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/rpcc"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func newRequestWillBeSentStream(logger zerolog.Logger, input network.RequestWillBeSentClient) runtime.Stream {
	return events.NewEventStream(input, func(_ context.Context, stream rpcc.Stream) (runtime.Value, error) {
		repl, err := stream.(network.RequestWillBeSentClient).Recv()

		if err != nil {
			logger.Trace().Err(err).Msg("failed to read data from request event stream")

			return runtime.None, err
		}

		var frameID string

		if repl.FrameID != nil {
			frameID = string(*repl.FrameID)
		}

		logger.Trace().
			Str("url", repl.Request.URL).
			Str("document_url", repl.DocumentURL).
			Str("frame_id", frameID).
			Interface("data", repl.Request).
			Msg("received request event")

		return toDriverRequest(repl.Request), nil
	})
}

func newResponseReceivedReader(logger zerolog.Logger, client *cdp.Client, input network.ResponseReceivedClient) runtime.Stream {
	return events.NewEventStream(input, func(ctx context.Context, stream rpcc.Stream) (runtime.Value, error) {
		repl, err := stream.(network.ResponseReceivedClient).Recv()

		if err != nil {
			logger.Trace().Err(err).Msg("failed to read data from request event stream")

			return runtime.None, err
		}

		var frameID string

		if repl.FrameID != nil {
			frameID = string(*repl.FrameID)
		}

		logger.Trace().
			Str("url", repl.Response.URL).
			Str("frame_id", frameID).
			Str("request_id", string(repl.RequestID)).
			Interface("data", repl.Response).
			Msg("received response event")

		var body []byte

		resp, err := client.Network.GetResponseBody(ctx, network.NewGetResponseBodyArgs(repl.RequestID))

		if err == nil {
			if resp.Base64Encoded {
				body, err = base64.StdEncoding.DecodeString(resp.Body)

				if err != nil {
					logger.Warn().
						Str("url", repl.Response.URL).
						Str("frame_id", frameID).
						Str("request_id", string(repl.RequestID)).
						Interface("data", repl.Response).
						Msg("failed to decode response body")
				}
			} else {
				body = []byte(resp.Body)
			}
		} else {
			logger.Warn().
				Str("url", repl.Response.URL).
				Str("frame_id", frameID).
				Str("request_id", string(repl.RequestID)).
				Interface("data", repl.Response).
				Msg("failed to get response body")
		}

		return toDriverResponse(repl.Response, body), nil
	})
}
