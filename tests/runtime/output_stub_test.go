package main

import (
	"context"

	"github.com/MontFerret/ferret/v2"
)

type outputStub struct {
	content    *ferret.Content
	collectErr error
	closeErr   error
	closed     bool
}

func (o *outputStub) Metadata() ferret.Metadata {
	return ferret.Metadata{}
}

func (o *outputStub) Consume(context.Context, ferret.Consumer) error {
	panic("expected collection")
}

func (o *outputStub) Collect(ctx context.Context) (*ferret.Content, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return o.content, o.collectErr
}

func (o *outputStub) Close() error {
	o.closed = true

	return o.closeErr
}
