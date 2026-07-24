package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Session creates a local session from a stateless model.
func Session(ctx context.Context, modelValue, optionsValue runtime.Value) (runtime.Value, error) {
	model, ok := modelValue.(core.Model)
	if !ok {
		return runtime.None, core.NewError(core.ErrInvalidOptions, "SESSION requires a stateless model")
	}

	options, err := core.DecodeSessionOptions(ctx, optionsValue)
	if err != nil {
		return runtime.None, err
	}

	return core.NewLocalSession(ctx, model, options)
}

// Reset clears the local history of a session.
func Reset(_ context.Context, value runtime.Value) (runtime.Value, error) {
	session, err := sessionValue(value)
	if err != nil {
		return runtime.None, err
	}

	if err := session.Reset(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}

// Fork returns an independent session copied from the current state.
func Fork(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	session, err := sessionValue(value)
	if err != nil {
		return runtime.None, err
	}

	return session.Fork(ctx)
}

func sessionValue(value runtime.Value) (core.Session, error) {
	session, ok := value.(core.Session)
	if !ok {
		return nil, core.NewError(core.ErrInvalidOptions, "expected an AI::LLM session")
	}

	return session, nil
}
