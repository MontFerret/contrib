package dom

import "context"

func withElementResult[T any](
	ctx context.Context,
	element *HTMLElement,
	operation func() (T, error),
) (T, error) {
	var zero T
	if err := element.ensureAttached(); err != nil {
		return zero, err
	}

	value, err := operation()
	if err != nil {
		return zero, element.normalizeError(ctx, err)
	}

	return value, nil
}

func withElementError(ctx context.Context, element *HTMLElement, operation func() error) error {
	_, err := withElementResult(ctx, element, func() (struct{}, error) {
		return struct{}{}, operation()
	})

	return err
}
