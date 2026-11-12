package dom

import (
	"context"
	"errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
)

func withDocumentResult[T any](
	ctx context.Context,
	doc *HTMLDocument,
	operation func(*documentState) (T, error),
) (T, error) {
	var zero T

	for attempt := 0; attempt < 2; attempt++ {
		state, err := doc.snapshot()
		if err != nil {
			return zero, err
		}

		if attempt == 0 && doc.rendererChanged(state) {
			if err := doc.refresh(ctx, state); err != nil {
				return zero, err
			}

			continue
		}

		result, err := operation(state)
		if err == nil {
			return result, nil
		}

		if errors.Is(err, drivers.ErrDetached) && !doc.isCurrentState(state) {
			continue
		}

		if attempt == 0 && (eval.IsStaleError(err) || doc.rendererChanged(state)) {
			if refreshErr := doc.refresh(ctx, state); refreshErr != nil {
				return zero, refreshErr
			}

			continue
		}

		return zero, err
	}

	return zero, drivers.ErrDetached
}

func withDocumentError(
	ctx context.Context,
	doc *HTMLDocument,
	operation func(*documentState) error,
) error {
	_, err := withDocumentResult(ctx, doc, func(state *documentState) (struct{}, error) {
		return struct{}{}, operation(state)
	})

	return err
}
