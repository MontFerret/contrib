package openai

import (
	"encoding/json"

	"github.com/openai/openai-go/v3/responses"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

func normalizeResponse(response *responses.Response) (core.Response, error) {
	if response == nil {
		return core.Response{}, core.NewError(core.ErrProvider, "provider returned no response")
	}

	if hasRefusal(response) {
		return core.Response{}, core.NewError(core.ErrRefusal, "provider refused the request")
	}

	switch response.Status {
	case responses.ResponseStatusCompleted:
	case responses.ResponseStatusIncomplete:
		if response.IncompleteDetails.Reason == "content_filter" {
			return core.Response{}, core.NewError(core.ErrRefusal, "provider refused the request")
		}

		return core.Response{}, core.NewError(core.ErrProvider, "provider returned an incomplete response")
	case responses.ResponseStatusFailed:
		return core.Response{}, normalizeResponseFailure(response.Error)
	default:
		return core.Response{}, core.NewError(core.ErrProvider, "provider did not complete the response")
	}

	raw := append(json.RawMessage(nil), []byte(response.RawJSON())...)

	return core.Response{
		ID:    response.ID,
		Model: response.Model,
		Text:  response.OutputText(),
		Usage: core.Usage{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
			TotalTokens:  response.Usage.TotalTokens,
		},
		RawJSON: raw,
	}, nil
}

func hasRefusal(response *responses.Response) bool {
	for _, output := range response.Output {
		for _, content := range output.Content {
			if content.Type == "refusal" {
				return true
			}
		}
	}

	return false
}

func normalizeResponseFailure(responseError responses.ResponseError) error {
	code := string(responseError.Code)

	switch code {
	case "rate_limit_exceeded":
		return core.NewError(core.ErrRateLimit, "provider rate limit exceeded")
	case "context_length_exceeded":
		return core.NewError(core.ErrContextLimit, "provider context limit exceeded")
	default:
		return core.NewError(core.ErrProvider, "provider failed to generate a response")
	}
}
