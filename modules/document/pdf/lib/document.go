package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// PageCount returns the number of pages in a PDF document.
func PageCount(_ context.Context, documentValue runtime.Value) (runtime.Value, error) {
	document, err := requireDocument(documentValue, "PAGE_COUNT")
	if err != nil {
		return runtime.None, err
	}

	count, err := document.PageCount()
	if err != nil {
		return runtime.None, err
	}

	return runtime.NewInt(count), nil
}

// Pages returns lazy page handles for a PDF document.
func Pages(ctx context.Context, documentValue runtime.Value) (runtime.Value, error) {
	document, err := requireDocument(documentValue, "PAGES")
	if err != nil {
		return runtime.None, err
	}

	pages, err := document.Pages()
	if err != nil {
		return runtime.None, err
	}

	out := runtime.NewArray(len(pages))
	for _, page := range pages {
		if err := out.Append(ctx, page); err != nil {
			return runtime.None, err
		}
	}

	return out, nil
}

// Page returns a lazy one-based page handle from a PDF document.
func Page(_ context.Context, documentValue, numberValue runtime.Value) (runtime.Value, error) {
	document, err := requireDocument(documentValue, "PAGE")
	if err != nil {
		return runtime.None, err
	}

	number, err := requireInt(numberValue, "PAGE", "number")
	if err != nil {
		return runtime.None, err
	}

	return document.Page(number)
}

// Text extracts best-effort text from a PDF document or page.
func Text(ctx context.Context, value runtime.Value) (runtime.Value, error) {
	switch typed := value.(type) {
	case *core.Document:
		text, err := typed.Text(ctx)
		if err != nil {
			return runtime.None, err
		}

		return runtime.NewString(text), nil
	case *core.Page:
		text, err := typed.Text(ctx)
		if err != nil {
			return runtime.None, err
		}

		return runtime.NewString(text), nil
	default:
		return runtime.None, fmt.Errorf("DOCUMENT::PDF TEXT failed: expected PDF document or page handle")
	}
}

// Close releases PDF document resources.
func Close(_ context.Context, documentValue runtime.Value) (runtime.Value, error) {
	document, err := requireDocument(documentValue, "CLOSE")
	if err != nil {
		return runtime.None, err
	}
	if err := document.Close(); err != nil {
		return runtime.None, err
	}

	return runtime.True, nil
}
