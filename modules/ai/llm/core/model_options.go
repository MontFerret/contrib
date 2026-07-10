package core

import (
	"context"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeModelOptions validates MODEL options without consulting process environment.
func DecodeModelOptions(ctx context.Context, value runtime.Value) (ModelOptions, error) {
	const label = "MODEL options"
	values, err := optionValues(ctx, value, label)
	if err != nil {
		return ModelOptions{}, err
	}

	allowed := map[string]struct{}{"model": {}, "apiKey": {}, "session": {}}
	if err := rejectUnknown(values, allowed, label); err != nil {
		return ModelOptions{}, err
	}

	model, found, err := stringOption(values, "model", label)
	if err != nil {
		return ModelOptions{}, err
	}
	if !found || strings.TrimSpace(model) == "" {
		return ModelOptions{}, NewError(ErrInvalidOptions, "MODEL options.model must not be blank")
	}

	apiKey, found, err := stringOption(values, "apiKey", label)
	if err != nil {
		return ModelOptions{}, err
	}
	if !found || strings.TrimSpace(apiKey) == "" {
		return ModelOptions{}, NewError(ErrInvalidOptions, "MODEL options.apiKey must not be blank")
	}

	session, _, err := boolOption(values, "session", label)
	if err != nil {
		return ModelOptions{}, err
	}

	return ModelOptions{Model: model, APIKey: apiKey, Session: session}, nil
}
