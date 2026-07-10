package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Generate produces text from a prompt.
func Generate(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return executeTextOperation(ctx, core.ModeGenerate, args...)
}

// Chat appends a user message and produces an assistant response.
func Chat(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return executeTextOperation(ctx, core.ModeChat, args...)
}

// Summarize produces a summary of the input text.
func Summarize(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return executeTextOperation(ctx, core.ModeSummarize, args...)
}

// Extract produces and validates a value against a JSON Schema.
func Extract(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 3, 4); err != nil {
		return runtime.None, err
	}

	target, input, err := targetAndInput(args)
	if err != nil {
		return runtime.None, err
	}

	schema, err := core.DecodeSchema(ctx, args[2])
	if err != nil {
		return runtime.None, err
	}

	semantic, execution, err := decodeOperationOptions(ctx, core.ModeExtract, args, 3)
	if err != nil {
		return runtime.None, err
	}
	semantic.Schema = schema

	return core.Execute(ctx, target, core.OperationRequest{
		Mode:      core.ModeExtract,
		Input:     input,
		Semantic:  semantic,
		Execution: execution,
	})
}

// Classify selects one label for the input text.
func Classify(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 3, 4); err != nil {
		return runtime.None, err
	}

	target, input, err := targetAndInput(args)
	if err != nil {
		return runtime.None, err
	}

	labels, err := core.DecodeLabels(ctx, args[2])
	if err != nil {
		return runtime.None, err
	}

	semantic, execution, err := decodeOperationOptions(ctx, core.ModeClassify, args, 3)
	if err != nil {
		return runtime.None, err
	}
	semantic.Labels = labels

	return core.Execute(ctx, target, core.OperationRequest{
		Mode:      core.ModeClassify,
		Input:     input,
		Semantic:  semantic,
		Execution: execution,
	})
}

func executeTextOperation(ctx context.Context, mode core.Mode, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 2, 3); err != nil {
		return runtime.None, err
	}

	target, input, err := targetAndInput(args)
	if err != nil {
		return runtime.None, err
	}

	semantic, execution, err := decodeOperationOptions(ctx, mode, args, 2)
	if err != nil {
		return runtime.None, err
	}

	return core.Execute(ctx, target, core.OperationRequest{
		Mode:      mode,
		Input:     input,
		Semantic:  semantic,
		Execution: execution,
	})
}

func targetAndInput(args []runtime.Value) (core.Target, string, error) {
	target, ok := args[0].(core.Target)
	if !ok {
		return nil, "", core.NewError(core.ErrInvalidOptions, "expected an AI::LLM model or session")
	}

	input, err := runtime.CastString(args[1])
	if err != nil {
		return nil, "", err
	}

	return target, input.String(), nil
}

func decodeOperationOptions(
	ctx context.Context,
	mode core.Mode,
	args []runtime.Value,
	index int,
) (core.SemanticOptions, core.ExecutionOptions, error) {
	if len(args) <= index {
		return core.SemanticOptions{}, core.ExecutionOptions{}, nil
	}

	return core.DecodeOperationOptions(ctx, mode, args[index])
}
