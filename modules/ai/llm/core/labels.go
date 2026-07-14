package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeLabels validates a nonempty array of unique, nonempty strings.
func DecodeLabels(ctx context.Context, value runtime.Value) ([]string, error) {
	list, ok := value.(runtime.List)
	if !ok {
		return nil, NewError(ErrInvalidOptions, "labels must be an array")
	}

	labels := make([]string, 0)
	seen := make(map[string]struct{})

	err := list.ForEach(ctx, func(_ context.Context, value runtime.Value, index runtime.Int) (runtime.Boolean, error) {
		label, ok := value.(runtime.String)
		if !ok {
			return runtime.False, NewError(ErrInvalidOptions, fmt.Sprintf("labels[%d] must be a string", index))
		}

		text := label.String()
		if text == "" {
			return runtime.False, NewError(ErrInvalidOptions, fmt.Sprintf("labels[%d] must not be empty", index))
		}

		if _, exists := seen[text]; exists {
			return runtime.False, NewError(ErrInvalidOptions, "labels must be unique")
		}

		seen[text] = struct{}{}
		labels = append(labels, text)

		return runtime.True, nil
	})

	if err != nil {
		return nil, err
	}

	if len(labels) == 0 {
		return nil, NewError(ErrInvalidOptions, "labels must not be empty")
	}

	return labels, nil
}
