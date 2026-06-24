package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func appendURLValues(ctx context.Context, values url.Values, owner string, input runtime.Value) error {
	if runtime.TypeNone.Is(input) {
		return nil
	}

	obj, err := requireMap(ctx, input, owner)
	if err != nil {
		return err
	}

	return obj.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		if err := appendURLValue(ctx, values, key.String(), value); err != nil {
			return runtime.False, fmt.Errorf("%s.%s: %w", owner, key.String(), err)
		}

		return runtime.True, nil
	})
}

func appendURLValue(ctx context.Context, values url.Values, key string, value runtime.Value) error {
	if runtime.TypeNone.Is(value) {
		return nil
	}

	if list, ok := value.(runtime.List); ok {
		iter, err := list.Iterate(ctx)
		if err != nil {
			return err
		}

		for {
			item, _, err := iter.Next(ctx)
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}
			if runtime.TypeNone.Is(item) {
				continue
			}

			values.Add(key, item.String())
		}

		return nil
	}

	if _, ok := value.(runtime.Map); ok {
		return fmt.Errorf("nested objects are not supported")
	}

	values.Add(key, value.String())

	return nil
}
