package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeSessionOptions validates v1 local-session settings.
func DecodeSessionOptions(ctx context.Context, value runtime.Value) (SessionOptions, error) {
	const label = "SESSION options"
	values, err := optionValues(ctx, value, label)
	if err != nil {
		return SessionOptions{}, err
	}

	if err := rejectUnknown(values, map[string]struct{}{"instructions": {}, "context": {}}, label); err != nil {
		return SessionOptions{}, err
	}

	options := SessionOptions{Context: ContextOptions{Mode: "local", Overflow: "error"}}

	if instructions, found, err := stringOption(values, "instructions", label); err != nil {
		return SessionOptions{}, err
	} else if found {
		options.Instructions = instructions
	}

	contextValue, found := values["context"]
	if !found {
		return options, nil
	}

	contextValues, err := optionValues(ctx, contextValue, "SESSION options.context")
	if err != nil {
		return SessionOptions{}, err
	}

	allowed := map[string]struct{}{"mode": {}, "overflow": {}, "maxTokens": {}, "reserveOutputTokens": {}}
	if err := rejectUnknown(contextValues, allowed, "SESSION options.context"); err != nil {
		return SessionOptions{}, err
	}

	if mode, present, err := stringOption(contextValues, "mode", "SESSION options.context"); err != nil {
		return SessionOptions{}, err
	} else if present {
		if mode != "local" {
			return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.mode must be local")
		}
		options.Context.Mode = mode
	}

	if overflow, present, err := stringOption(contextValues, "overflow", "SESSION options.context"); err != nil {
		return SessionOptions{}, err
	} else if present {
		if overflow != "error" {
			return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.overflow must be error")
		}
		options.Context.Overflow = overflow
	}

	maxTokens, hasMaxTokens, err := intOption(contextValues, "maxTokens", "SESSION options.context")
	if err != nil {
		return SessionOptions{}, err
	}

	if hasMaxTokens && maxTokens <= 0 {
		return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.maxTokens must be positive")
	}

	options.Context.MaxTokens = maxTokens

	reserve, hasReserve, err := intOption(contextValues, "reserveOutputTokens", "SESSION options.context")
	if err != nil {
		return SessionOptions{}, err
	}

	if hasReserve && reserve < 0 {
		return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.reserveOutputTokens must be nonnegative")
	}

	if hasMaxTokens && hasReserve && reserve >= maxTokens {
		return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION output token reserve must be smaller than maxTokens")
	}

	options.Context.ReserveOutputTokens = reserve

	return options, nil
}
