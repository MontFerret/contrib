package core

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeLabels validates a nonempty array of unique, nonempty strings.
func DecodeLabels(ctx context.Context, value runtime.Value) ([]string, error) {
	labels, err := decodeList[[]string](ctx, value, "labels")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})

	for index, label := range labels {
		if label == "" {
			return nil, NewError(ErrInvalidOptions, fmt.Sprintf("labels[%d] must not be empty", index))
		}

		if _, exists := seen[label]; exists {
			return nil, NewError(ErrInvalidOptions, "labels must be unique")
		}

		seen[label] = struct{}{}
	}

	if len(labels) == 0 {
		return nil, NewError(ErrInvalidOptions, "labels must not be empty")
	}

	return labels, nil
}
