package common

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// CollectFrames recursively collects all frames from the given document and appends them to the receiver list.
// It first appends the current document to the receiver list, then retrieves its child documents and calls itself for each child.
func CollectFrames(ctx context.Context, receiver runtime.List, doc drivers.HTMLDocument) error {
	err := receiver.Append(ctx, doc)

	if err != nil {
		return err
	}

	children, err := doc.GetChildDocuments(ctx)

	if err != nil {
		return err
	}

	return children.ForEach(ctx, func(ctx context.Context, value runtime.Value, idx runtime.Int) (runtime.Boolean, error) {
		childDoc, err := CastHTMLDocument(value)

		if err != nil {
			return false, err
		}

		err = CollectFrames(ctx, receiver, childDoc)

		if err != nil {
			return false, err
		}

		return true, nil
	})
}
