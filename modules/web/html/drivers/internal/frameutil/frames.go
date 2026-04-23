package frameutil

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func Collect(ctx context.Context, receiver runtime.List, doc drivers.HTMLDocument) error {
	if err := receiver.Append(ctx, doc); err != nil {
		return err
	}

	children, err := doc.GetChildDocuments(ctx)
	if err != nil {
		return err
	}

	return children.ForEach(ctx, func(ctx context.Context, value runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
		childDoc, err := drivers.ToDocument(value)
		if err != nil {
			return false, err
		}

		if err := Collect(ctx, receiver, childDoc); err != nil {
			return false, err
		}

		return true, nil
	})
}
