package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Generate produces text from a prompt.
//
// @param target {Model|Session} Model or session to execute against.
// @param prompt {String} User prompt.
// @param options {Object?} Generation and execution options.
// @return {String} Generated text.
func Generate(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return executeTextOperation(ctx, core.ModeGenerate, args...)
}

// Chat produces an assistant response to a user message.
//
// A successful session call commits the supplied messages and response to the
// session history.
//
// @param target {Model|Session} Model or session to execute against.
// @param message {String} Final user message.
// @param options {Object?} Messages, instructions, and execution options.
// @return {String} Assistant response.
func Chat(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return executeTextOperation(ctx, core.ModeChat, args...)
}

// Summarize produces a summary of the input text.
//
// @param target {Model|Session} Model or session to execute against.
// @param text {String} Text to summarize.
// @param options {Object?} Summarization and execution options.
// @return {String} Generated summary.
func Summarize(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	return executeTextOperation(ctx, core.ModeSummarize, args...)
}

// Extract produces structured data and validates it against a JSON Schema.
//
// @param target {Model|Session} Model or session to execute against.
// @param text {String} Source text.
// @param schema {Object} JSON Schema for the result.
// @param options {Object?} Extraction and execution options.
// @return {Any} Validated structured value.
func Extract(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 3, 4); err != nil {
		return runtime.None, err
	}

	target, input, err := targetAndInput(ctx, args)
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

// Classify selects one allowed label for the input text.
//
// @param target {Model|Session} Model or session to execute against.
// @param text {String} Text to classify.
// @param labels {Array<String>} Allowed labels.
// @param options {Object?} Classification and execution options.
// @return {Object} Object containing the selected label.
func Classify(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 3, 4); err != nil {
		return runtime.None, err
	}

	target, input, err := targetAndInput(ctx, args)
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

	target, input, err := targetAndInput(ctx, args)
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

func targetAndInput(ctx context.Context, args []runtime.Value) (core.Target, string, error) {
	target, ok := args[0].(core.Target)
	if !ok {
		return nil, "", core.NewError(core.ErrInvalidOptions, "expected an AI::LLM model or session")
	}

	input, err := sdk.DecodeArg[string](
		ctx,
		args,
		1,
		sdk.RequireType(runtime.TypeString),
	)
	if err != nil {
		return nil, "", err
	}

	return target, input, nil
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
