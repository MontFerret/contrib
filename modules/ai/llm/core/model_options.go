package core

import (
	"context"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type modelOptionsInput struct {
	Model   string `json:"model"`
	APIKey  string `json:"apiKey"`
	Session bool   `json:"session"`
}

// DecodeModelOptions validates MODEL options without consulting process environment.
func DecodeModelOptions(ctx context.Context, value runtime.Value) (ModelOptions, error) {
	const label = "MODEL options"
	input, err := decodeOptionObject[modelOptionsInput](ctx, value, label)
	if err != nil {
		return ModelOptions{}, err
	}

	if strings.TrimSpace(input.Model) == "" {
		return ModelOptions{}, NewError(ErrInvalidOptions, "MODEL options.model must not be blank")
	}

	if strings.TrimSpace(input.APIKey) == "" {
		return ModelOptions{}, NewError(ErrInvalidOptions, "MODEL options.apiKey must not be blank")
	}

	return ModelOptions(input), nil
}
