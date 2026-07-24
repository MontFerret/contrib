package core

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	sessionOptionsInput struct {
		Context      runtime.Value `json:"context"`
		Instructions string        `json:"instructions"`
	}

	sessionContextOptionsInput struct {
		Mode                *string `json:"mode"`
		Overflow            *string `json:"overflow"`
		MaxTokens           *int64  `json:"maxTokens"`
		ReserveOutputTokens *int64  `json:"reserveOutputTokens"`
	}
)

// DecodeSessionOptions validates v1 local-session settings.
func DecodeSessionOptions(ctx context.Context, value runtime.Value) (SessionOptions, error) {
	const label = "SESSION options"
	input, err := decodeOptionObject[sessionOptionsInput](ctx, value, label)
	if err != nil {
		return SessionOptions{}, err
	}

	options := SessionOptions{
		Instructions: input.Instructions,
		Context: ContextOptions{
			Mode:     "local",
			Overflow: "error",
		},
	}

	if input.Context == nil || runtime.TypeNone.Is(input.Context) {
		return options, nil
	}

	const contextLabel = "SESSION options.context"
	contextInput, err := decodeOptionObject[sessionContextOptionsInput](ctx, input.Context, contextLabel)
	if err != nil {
		return SessionOptions{}, err
	}

	if contextInput.Mode != nil {
		if *contextInput.Mode != "local" {
			return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.mode must be local")
		}
		options.Context.Mode = *contextInput.Mode
	}

	if contextInput.Overflow != nil {
		if *contextInput.Overflow != "error" {
			return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.overflow must be error")
		}
		options.Context.Overflow = *contextInput.Overflow
	}

	if contextInput.MaxTokens != nil {
		if *contextInput.MaxTokens <= 0 {
			return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.maxTokens must be positive")
		}
		options.Context.MaxTokens = *contextInput.MaxTokens
	}

	if contextInput.ReserveOutputTokens != nil {
		if *contextInput.ReserveOutputTokens < 0 {
			return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION options.context.reserveOutputTokens must be nonnegative")
		}
		options.Context.ReserveOutputTokens = *contextInput.ReserveOutputTokens
	}

	if contextInput.MaxTokens != nil &&
		contextInput.ReserveOutputTokens != nil &&
		*contextInput.ReserveOutputTokens >= *contextInput.MaxTokens {
		return SessionOptions{}, NewError(ErrInvalidOptions, "SESSION output token reserve must be smaller than maxTokens")
	}

	return options, nil
}
