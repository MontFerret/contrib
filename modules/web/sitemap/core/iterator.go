package core

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// URLIterator lazily expands sitemap documents into URL entries.
type URLIterator struct {
	walker *walker
	count  runtime.Int
}

// NewURLIterator creates a lazy sitemap URL iterator.
func NewURLIterator(target string, opts Options) *URLIterator {
	return &URLIterator{
		walker: newWalker(target, opts),
	}
}

// Iterate returns the iterator itself.
func (i *URLIterator) Iterate(_ context.Context) (runtime.Iterator, error) {
	return i, nil
}

// Next returns the next URL entry and its 1-based yield index.
func (i *URLIterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	entry, err := i.walker.nextEntry(ctx)
	if err != nil {
		if err == io.EOF {
			return runtime.None, runtime.None, io.EOF
		}

		return runtime.None, runtime.None, err
	}

	i.count++

	return entry.ToValue(), i.count, nil
}

// Close stops iteration.
func (i *URLIterator) Close() error {
	i.walker.close()

	return nil
}
