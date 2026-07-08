package core

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// PageIterator lazily creates page values while iterating a page collection.
type PageIterator struct {
	collection *PageCollection
	pos        runtime.Int
}

// NewPageIterator creates an iterator for a lazy page collection.
func NewPageIterator(collection *PageCollection) *PageIterator {
	return &PageIterator{collection: collection}
}

func (it *PageIterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	value, found, err := it.collection.LookupAt(ctx, it.pos)
	if err != nil {
		return runtime.None, runtime.None, err
	}
	if !found {
		return runtime.None, runtime.None, io.EOF
	}

	key := it.pos
	it.pos++

	return value, key, nil
}
