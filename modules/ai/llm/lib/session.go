package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Session creates a local conversation session from a stateless model.
//
// @param model {Model} Stateless model.
// @param options {Object} Session configuration.
// @return {Session} Local conversation session.
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

// Reset clears the visible history of a session.
//
// @param session {Session} Session to reset.
// @return {Boolean} True when the history is cleared.
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

// Fork creates an independent session with copied state and history.
//
// @param session {Session} Session to copy.
// @return {Session} Independent copied session.
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
