package common

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type Iterator struct {
	node drivers.HTMLElement
	pos  runtime.Int
}

func NewIterator(
	node drivers.HTMLElement,
) (runtime.Iterator, error) {
	if node == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "result")
	}

	return &Iterator{node, 0}, nil
}

func (iter *Iterator) HasNext(ctx context.Context) (bool, error) {
	size, err := iter.node.Length(ctx)

	if err != nil {
		return false, fmt.Errorf("failed to get length of the node: %w", err)
	}

	return size > iter.pos, nil
}

func (iter *Iterator) Next(ctx context.Context) (runtime.Value, runtime.Value, error) {
	idx := iter.pos
	val, err := iter.node.GetChildNode(ctx, idx)

	if err != nil {
		return runtime.None, runtime.None, err
	}

	iter.pos++

	return val, idx, nil
}
