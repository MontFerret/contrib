package core

import (
	"context"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// PageCollection is a lazy, zero-indexed view of a PDF document's pages.
type PageCollection struct {
	document *Document
	id       uint64
}

// NewPageCollection creates a lazy page collection for a document.
func NewPageCollection(document *Document) *PageCollection {
	return &PageCollection{
		document: document,
		id:       newResourceID(),
	}
}

func (c *PageCollection) At(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	value, _, err := c.lookupAt(ctx, idx)

	return value, err
}

func (c *PageCollection) LookupAt(ctx context.Context, idx runtime.Int) (runtime.Value, bool, error) {
	return c.lookupAt(ctx, idx)
}

func (c *PageCollection) Iterate(ctx context.Context) (runtime.Iterator, error) {
	if err := c.ensureOpen(ctx); err != nil {
		return nil, err
	}

	return NewPageIterator(c), nil
}

func (c *PageCollection) Length(ctx context.Context) (runtime.Int, error) {
	if err := ctx.Err(); err != nil {
		return runtime.ZeroInt, OperationError("PAGES", err)
	}

	count, err := c.document.PageCount()
	if err != nil {
		return runtime.ZeroInt, err
	}

	return runtime.NewInt(count), nil
}

func (c *PageCollection) String() string {
	return commonresource.Display("document.pdf.pages")
}

func (c *PageCollection) Hash() uint64 {
	return commonresource.Hash("document.pdf.pages", c.id)
}

func (c *PageCollection) Copy() runtime.Value {
	return c
}

func (c *PageCollection) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalDisplayJSON("document.pdf.pages")
}

func (c *PageCollection) lookupAt(ctx context.Context, idx runtime.Int) (runtime.Value, bool, error) {
	if err := ctx.Err(); err != nil {
		return runtime.None, false, OperationError("PAGES", err)
	}
	if idx < 0 {
		return runtime.None, false, nil
	}

	count, err := c.document.PageCount()
	if err != nil {
		return runtime.None, false, err
	}
	if idx >= runtime.Int(count) {
		return runtime.None, false, nil
	}

	page, err := c.document.Page(int(idx) + 1)
	if err != nil {
		return runtime.None, false, err
	}

	return page, true, nil
}

func (c *PageCollection) ensureOpen(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return OperationError("PAGES", err)
	}

	c.document.mu.RLock()
	defer c.document.mu.RUnlock()

	if err := c.document.ensureOpen(); err != nil {
		return OperationError("PAGES", err)
	}

	return nil
}
