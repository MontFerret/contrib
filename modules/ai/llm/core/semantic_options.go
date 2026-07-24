package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type semanticExecutionOptionsInput struct {
	Schema          runtime.Value `json:"schema"`
	Messages        runtime.Value `json:"messages"`
	Labels          runtime.Value `json:"labels"`
	MaxWords        *int64        `json:"maxWords"`
	Temperature     *float64      `json:"temperature"`
	MaxOutputTokens *int64        `json:"maxOutputTokens"`
	Timeout         *int64        `json:"timeout"`
	Instructions    string        `json:"instructions"`
	Style           string        `json:"style"`
}

// DecodeSemanticOptions validates mode-specific semantic options.
func DecodeSemanticOptions(ctx context.Context, mode Mode, value runtime.Value) (SemanticOptions, error) {
	label := string(mode) + " options"
	value, err := snapshotOptionObject(ctx, value, label)
	if err != nil {
		return SemanticOptions{}, err
	}

	allowed, err := semanticKeys(mode)
	if err != nil {
		return SemanticOptions{}, err
	}

	input, err := decodeOptionObject[semanticExecutionOptionsInput](
		ctx,
		value,
		label,
		sdk.OnlyFields(allowed...),
	)
	if err != nil {
		return SemanticOptions{}, err
	}

	return decodeSemanticInput(ctx, mode, input)
}

// DecodeOperationOptions validates combined function semantic and execution options.
func DecodeOperationOptions(ctx context.Context, mode Mode, value runtime.Value) (SemanticOptions, ExecutionOptions, error) {
	const label = "function options"
	value, err := snapshotOptionObject(ctx, value, label)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	semanticAllowed, err := functionSemanticKeys(mode)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	allowed := append(semanticAllowed, "temperature", "maxOutputTokens", "timeout")
	input, err := decodeOptionObject[semanticExecutionOptionsInput](
		ctx,
		value,
		label,
		sdk.OnlyFields(allowed...),
	)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	semantic, err := decodeSemanticInput(ctx, mode, input)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	execution, err := decodeExecutionInput(executionOptionsInput{
		Temperature:     input.Temperature,
		MaxOutputTokens: input.MaxOutputTokens,
		Timeout:         input.Timeout,
	}, label)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	return semantic, execution, nil
}

func functionSemanticKeys(mode Mode) ([]string, error) {
	allowed, err := semanticKeys(mode)
	if err != nil {
		return nil, err
	}

	switch mode {
	case ModeExtract:
		return []string{"instructions"}, nil
	case ModeClassify:
		return []string{"instructions"}, nil
	default:
		return allowed, nil
	}
}

func semanticKeys(mode Mode) ([]string, error) {
	switch mode {
	case ModeGenerate:
		return []string{"instructions"}, nil
	case ModeChat:
		return []string{"messages", "instructions"}, nil
	case ModeSummarize:
		return []string{"style", "maxWords", "instructions"}, nil
	case ModeExtract:
		return []string{"schema", "instructions"}, nil
	case ModeClassify:
		return []string{"labels", "instructions"}, nil
	default:
		return nil, NewError(ErrUnsupportedOperation, "unsupported AI::LLM operation")
	}
}

func decodeSemanticInput(ctx context.Context, mode Mode, input semanticExecutionOptionsInput) (SemanticOptions, error) {
	options := SemanticOptions{
		Instructions: input.Instructions,
	}

	switch mode {
	case ModeGenerate:
	case ModeChat:
		if input.Messages != nil {
			messages, err := DecodeMessages(ctx, input.Messages)
			if err != nil {
				return options, err
			}

			options.Messages = messages
		}
	case ModeSummarize:
		options.Style = input.Style
		if input.MaxWords != nil {
			if *input.MaxWords <= 0 {
				return options, NewError(ErrInvalidOptions, "summarize options.maxWords must be positive")
			}

			options.MaxWords = *input.MaxWords
		}
	case ModeExtract:
		if input.Schema != nil {
			schema, err := DecodeSchema(ctx, input.Schema)
			if err != nil {
				return options, err
			}

			options.Schema = schema
		}
	case ModeClassify:
		if input.Labels != nil {
			labels, err := DecodeLabels(ctx, input.Labels)
			if err != nil {
				return options, err
			}

			options.Labels = labels
		}
	default:
		return options, NewError(ErrUnsupportedOperation, "unsupported AI::LLM operation")
	}

	return options, nil
}
