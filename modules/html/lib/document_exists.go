package html

import (
	"context"
	"net/http"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// DOCUMENT_EXISTS returns a boolean value indicating whether a web page exists by a given url.
// @param {String} url - Target url.
// @param {Object} [options] - Request options.
// @param {Object} [options.headers] - Request headers.
// @return {Boolean} - A boolean value indicating whether a web page exists by a given url.
func DocumentExists(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return nil, err
	}

	if err := runtime.ValidateType(args[0], runtime.TypeString); err != nil {
		return nil, err
	}

	url := args[0].String()

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return runtime.None, err
	}

	if len(args) > 1 {
		options, err := runtime.CastMap(args[1])

		if err != nil {
			return nil, err
		}

		headers, exist, err := sdk.TryGetByKey[runtime.Map](ctx, options, runtime.String("headers"))

		if err != nil {
			return nil, err
		}

		if exist {
			req.Header = http.Header{}

			err = headers.ForEach(ctx, func(c context.Context, value, key runtime.Value) (runtime.Boolean, error) {
				req.Header.Set(key.String(), value.String())

				return true, nil
			})

			if err != nil {
				return nil, err
			}
		}
	}

	resp, err := client.Do(req.WithContext(ctx))

	if err != nil {
		return runtime.False, nil
	}

	var exists runtime.Boolean

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		exists = runtime.True
	}

	return exists, nil
}
