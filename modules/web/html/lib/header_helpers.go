package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func parseHeadersValue(ctx context.Context, value runtime.Value) (*drivers.HTTPHeaders, error) {
	if value == nil || value == runtime.None {
		return nil, nil
	}

	if headers, ok := value.(*drivers.HTTPHeaders); ok {
		return headers, nil
	}

	if headers, ok := runtime.UnwrapAs[*drivers.HTTPHeaders](value); ok {
		return headers, nil
	}

	values, err := sdk.DecodeValue[map[string]string](ctx, value, sdk.DisallowNoneValues())
	if err != nil {
		return nil, fmt.Errorf("headers: %w", err)
	}

	headers := drivers.NewHTTPHeaders()

	for key, val := range values {
		if err := headers.Set(ctx, runtime.String(key), runtime.String(val)); err != nil {
			return nil, fmt.Errorf("headers: %w", err)
		}
	}

	return headers, nil
}
