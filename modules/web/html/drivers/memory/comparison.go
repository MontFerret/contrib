package memory

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func compareStrings(first, second string) runtime.Ordering {
	if first < second {
		return runtime.Less
	}

	if first > second {
		return runtime.Greater
	}

	return runtime.Equal
}

func invalidComparison(first, second runtime.Value) (runtime.Ordering, error) {
	return runtime.Equal, runtime.Errorf(
		runtime.ErrInvalidOperation,
		"cannot compare %s with %s",
		runtime.TypeName(runtime.TypeOf(first)),
		runtime.TypeName(runtime.TypeOf(second)),
	)
}

func checkComparisonContext(ctx context.Context) (runtime.Ordering, error) {
	if err := ctx.Err(); err != nil {
		return runtime.Equal, err
	}

	return runtime.Equal, nil
}
