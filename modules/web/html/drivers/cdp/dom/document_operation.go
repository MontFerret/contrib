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

		result, err := operation(state)
		if err == nil {
			return result, nil
		}

		if errors.Is(err, drivers.ErrDetached) && doc.generationChanged(state.generation) {
			continue
		}

		if attempt == 0 && eval.IsStaleError(err) {
			if refreshErr := doc.refresh(ctx, state.generation); refreshErr != nil {
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
