package openai

import (
	"context"
	"time"

	sdkopenai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

func buildRequest(model string, request core.Request) (responses.ResponseNewParams, error) {
	input, err := mapMessages(request.Messages)
	if err != nil {
		return responses.ResponseNewParams{}, err
	}

	params := responses.ResponseNewParams{
		Model:      model,
		Input:      responses.ResponseNewParamsInputUnion{OfInputItemList: input},
		Store:      sdkopenai.Bool(false),
		Truncation: responses.ResponseNewParamsTruncationDisabled,
	}

	if request.Instructions != "" {
		params.Instructions = sdkopenai.String(request.Instructions)
	}
	if request.Options.Temperature != nil {
		params.Temperature = sdkopenai.Float(*request.Options.Temperature)
	}
	if request.Options.MaxOutputTokens > 0 {
		params.MaxOutputTokens = sdkopenai.Int(request.Options.MaxOutputTokens)
	}

	return params, nil
}

func buildStructuredRequest(model string, request core.StructuredRequest) (responses.ResponseNewParams, error) {
	params, err := buildRequest(model, request.Request)
	if err != nil {
		return responses.ResponseNewParams{}, err
	}

	document := request.Schema.Document()

	if err := validateStructuredOutputSchema(document); err != nil {
		return responses.ResponseNewParams{}, err
	}

	format := &responses.ResponseFormatTextJSONSchemaConfigParam{
		Name:   request.Name,
		Schema: document,
		Strict: sdkopenai.Bool(true),
	}

	if request.Description != "" {
		format.Description = sdkopenai.String(request.Description)
	}

	params.Text = responses.ResponseTextConfigParam{
		Format: responses.ResponseFormatTextConfigUnionParam{OfJSONSchema: format},
	}

	return params, nil
}

func mapMessages(messages []core.Message) (responses.ResponseInputParam, error) {
	input := make(responses.ResponseInputParam, 0, len(messages))

	for _, message := range messages {
		if message.Content.Type != core.ContentText {
			return nil, core.NewError(core.ErrInvalidOptions, "only text message content is supported")
		}

		role, err := mapRole(message.Role)
		if err != nil {
			return nil, err
		}

		input = append(input, responses.ResponseInputItemParamOfMessage(message.Content.Text, role))
	}

	return input, nil
}

func mapRole(role core.Role) (responses.EasyInputMessageRole, error) {
	switch role {
	case core.RoleSystem:
		return responses.EasyInputMessageRoleSystem, nil
	case core.RoleDeveloper:
		return responses.EasyInputMessageRoleDeveloper, nil
	case core.RoleUser:
		return responses.EasyInputMessageRoleUser, nil
	case core.RoleAssistant:
		return responses.EasyInputMessageRoleAssistant, nil
	default:
		return "", core.NewError(core.ErrInvalidOptions, "unsupported message role")
	}
}

func requestContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}
