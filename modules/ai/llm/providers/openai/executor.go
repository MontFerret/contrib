package openai

import (
	"context"

	sdkopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

type executor struct {
	model  string
	client sdkopenai.Client
}

func newExecutor(model, apiKey string) *executor {
	client := sdkopenai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(openAIBaseURL),
		option.WithHTTPClient(ferretHTTPClient{}),
		option.WithMaxRetries(0),
	)

	return &executor{client: client, model: model}
}

func (e *executor) Generate(ctx context.Context, request core.Request) (core.Response, error) {
	params, err := buildRequest(e.model, request)
	if err != nil {
		return core.Response{}, err
	}

	requestCtx, cancel := requestContext(ctx, request.Options.Timeout)
	defer cancel()

	response, err := e.client.Responses.New(requestCtx, params)
	if err != nil {
		return core.Response{}, normalizeError(err)
	}

	return normalizeResponse(response)
}

func (e *executor) GenerateStructured(ctx context.Context, request core.StructuredRequest) (core.Response, error) {
	params, err := buildStructuredRequest(e.model, request)
	if err != nil {
		return core.Response{}, err
	}

	requestCtx, cancel := requestContext(ctx, request.Options.Timeout)
	defer cancel()

	response, err := e.client.Responses.New(requestCtx, params)
	if err != nil {
		return core.Response{}, normalizeError(err)
	}

	return normalizeResponse(response)
}
