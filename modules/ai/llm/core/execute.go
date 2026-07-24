package core

import (
	"context"
	"errors"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type transactionalTarget interface {
	executeOperation(context.Context, OperationRequest) (runtime.Value, error)
}

// Execute routes every function and Queryable operation through one dispatcher.
func Execute(ctx context.Context, target Target, operation OperationRequest) (runtime.Value, error) {
	if target == nil {
		return runtime.None, NewError(ErrInvalidOptions, "model or session is required")
	}

	if session, ok := target.(transactionalTarget); ok {
		return session.executeOperation(ctx, operation)
	}

	value, _, _, err := executeAgainst(ctx, target, operation, nil, "")
	if err != nil {
		return runtime.None, OperationError(strings.ToUpper(string(operation.Mode)), err)
	}

	return value, nil
}

func executeAgainst(
	ctx context.Context,
	target Target,
	operation OperationRequest,
	history []Message,
	persistentInstructions string,
) (runtime.Value, Response, []Message, error) {
	requestCtx, cancel, err := executionContext(ctx, operation.Execution)
	if err != nil {
		return runtime.None, Response{}, nil, err
	}

	defer cancel()

	switch operation.Mode {
	case ModeGenerate, ModeChat, ModeSummarize:
		request, err := BuildGenerationRequest(operation)
		if err != nil {
			return runtime.None, Response{}, nil, err
		}

		inputs := copyMessages(request.Messages)
		request.Messages = append(copyMessages(history), request.Messages...)
		request.Instructions = joinInstructions(persistentInstructions, request.Instructions)
		response, err := target.Generate(requestCtx, request)

		if err != nil {
			return runtime.None, Response{}, nil, normalizeContextError(requestCtx, err)
		}

		return runtime.NewString(response.Text), response, inputs, nil
	case ModeExtract, ModeClassify:
		request, err := BuildStructuredRequest(operation)
		if err != nil {
			return runtime.None, Response{}, nil, err
		}

		inputs := copyMessages(request.Messages)
		request.Messages = append(copyMessages(history), request.Messages...)
		request.Instructions = joinInstructions(persistentInstructions, request.Instructions)
		response, err := target.GenerateStructured(requestCtx, request)

		if err != nil {
			return runtime.None, Response{}, nil, normalizeContextError(requestCtx, err)
		}

		value, err := request.Schema.ValidateJSON([]byte(response.Text))
		if err != nil {
			return runtime.None, Response{}, nil, err
		}

		return value, response, inputs, nil
	default:
		return runtime.None, Response{}, nil, NewError(ErrUnsupportedOperation, "unsupported AI::LLM operation")
	}
}

func executionContext(ctx context.Context, options ExecutionOptions) (context.Context, context.CancelFunc, error) {
	if options.Timeout < 0 {
		return nil, nil, NewError(ErrInvalidOptions, "timeout must be nonnegative")
	}

	if options.Timeout == 0 {
		return ctx, func() {}, nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, options.Timeout)

	return requestCtx, cancel, nil
}

func normalizeContextError(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return NewError(ErrTimeout, "provider request timed out")
	}

	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(err, context.Canceled) {
		return context.Canceled
	}

	return err
}
