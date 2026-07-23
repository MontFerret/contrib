package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeSemanticOptions validates mode-specific semantic options.
func DecodeSemanticOptions(ctx context.Context, mode Mode, value runtime.Value) (SemanticOptions, error) {
	values, err := optionValues(ctx, value, string(mode)+" options")
	if err != nil {
		return SemanticOptions{}, err
	}

	allowed, err := semanticKeys(mode)
	if err != nil {
		return SemanticOptions{}, err
	}

	if err := rejectUnknown(values, allowed, string(mode)+" options"); err != nil {
		return SemanticOptions{}, err
	}

	return decodeSemanticValues(ctx, mode, values)
}

// DecodeOperationOptions validates combined function semantic and execution options.
func DecodeOperationOptions(ctx context.Context, mode Mode, value runtime.Value) (SemanticOptions, ExecutionOptions, error) {
	const label = "function options"
	values, err := optionValues(ctx, value, label)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	semanticAllowed, err := functionSemanticKeys(mode)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	allAllowed := make(map[string]struct{}, len(semanticAllowed)+3)

	for key := range semanticAllowed {
		allAllowed[key] = struct{}{}
	}

	for _, key := range []string{"temperature", "maxOutputTokens", "timeout"} {
		allAllowed[key] = struct{}{}
	}

	if err := rejectUnknown(values, allAllowed, label); err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	semanticValues := make(map[string]runtime.Value, len(semanticAllowed))
	for key := range semanticAllowed {
		if current, found := values[key]; found {
			semanticValues[key] = current
		}
	}

	executionValues := make(map[string]runtime.Value, 3)
	for _, key := range []string{"temperature", "maxOutputTokens", "timeout"} {
		if current, found := values[key]; found {
			executionValues[key] = current
		}
	}

	semantic, err := decodeSemanticValues(ctx, mode, semanticValues)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	execution, err := decodeExecutionValues(executionValues, label)
	if err != nil {
		return SemanticOptions{}, ExecutionOptions{}, err
	}

	return semantic, execution, nil
}

func functionSemanticKeys(mode Mode) (map[string]struct{}, error) {
	allowed, err := semanticKeys(mode)
	if err != nil {
		return nil, err
	}

	copy := make(map[string]struct{}, len(allowed))
	for key := range allowed {
		copy[key] = struct{}{}
	}

	delete(copy, "schema")
	delete(copy, "labels")

	return copy, nil
}

func semanticKeys(mode Mode) (map[string]struct{}, error) {
	switch mode {
	case ModeGenerate:
		return map[string]struct{}{"instructions": {}}, nil
	case ModeChat:
		return map[string]struct{}{"messages": {}, "instructions": {}}, nil
	case ModeSummarize:
		return map[string]struct{}{"style": {}, "maxWords": {}, "instructions": {}}, nil
	case ModeExtract:
		return map[string]struct{}{"schema": {}, "instructions": {}}, nil
	case ModeClassify:
		return map[string]struct{}{"labels": {}, "instructions": {}}, nil
	default:
		return nil, NewError(ErrUnsupportedOperation, "unsupported AI::LLM operation")
	}
}

func decodeSemanticValues(ctx context.Context, mode Mode, values map[string]runtime.Value) (SemanticOptions, error) {
	var options SemanticOptions
	if instructions, found, err := stringOption(values, "instructions", string(mode)+" options"); err != nil {
		return options, err
	} else if found {
		options.Instructions = instructions
	}

	switch mode {
	case ModeGenerate:
	case ModeChat:
		if value, found := values["messages"]; found {
			messages, err := DecodeMessages(ctx, value)
			if err != nil {
				return options, err
			}

			options.Messages = messages
		}
	case ModeSummarize:
		if style, found, err := stringOption(values, "style", "summarize options"); err != nil {
			return options, err
		} else if found {
			options.Style = style
		}

		if maxWords, found, err := intOption(values, "maxWords", "summarize options"); err != nil {
			return options, err
		} else if found {
			if maxWords <= 0 {
				return options, NewError(ErrInvalidOptions, "summarize options.maxWords must be positive")
			}

			options.MaxWords = maxWords
		}
	case ModeExtract:
		if value, found := values["schema"]; found {
			schema, err := DecodeSchema(ctx, value)
			if err != nil {
				return options, err
			}

			options.Schema = schema
		}
	case ModeClassify:
		if value, found := values["labels"]; found {
			labels, err := DecodeLabels(ctx, value)
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
