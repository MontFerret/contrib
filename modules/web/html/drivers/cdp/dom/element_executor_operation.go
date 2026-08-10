package dom

import "context"

func runElementResult[T any](
	ctx context.Context,
	executor *elementExecutor,
	zeroValue func() T,
	operation func() (T, error),
) (T, error) {
	if err := executor.ensureAttached(); err != nil {
		return zeroValue(), err
	}

	value, err := operation()
	if err != nil {
		return zeroValue(), executor.normalizeError(ctx, err)
	}

	return value, nil
}
