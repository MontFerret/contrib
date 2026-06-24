package core

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func decodeHeaders(ctx context.Context, value runtime.Value, owner string) (http.Header, error) {
	obj, err := requireMap(ctx, value, owner)
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	err = obj.ForEach(ctx, func(ctx context.Context, value, key runtime.Value) (runtime.Boolean, error) {
		if runtime.TypeNone.Is(value) {
			return runtime.True, nil
		}

		if list, ok := value.(runtime.List); ok {
			return runtime.True, list.ForEach(ctx, func(ctx context.Context, item runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
				if runtime.TypeNone.Is(item) {
					return runtime.True, nil
				}

				headers.Add(key.String(), item.String())

				return runtime.True, nil
			})
		}

		headers.Set(key.String(), value.String())

		return runtime.True, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", owner, err)
	}

	return headers, nil
}

func mergeHeaders(base, override http.Header) http.Header {
	out := base.Clone()

	for key, values := range override {
		delete(out, key)

		for _, value := range values {
			out.Add(key, value)
		}
	}

	return out
}

func hasHeader(headers http.Header, key string) bool {
	_, found := headers[http.CanonicalHeaderKey(key)]

	return found
}
