package events

import (
	"context"

	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	Decoder func(ctx context.Context, stream rpcc.Stream) (runtime.Value, error)

	Factory func(ctx context.Context) (rpcc.Stream, error)

	EventStream struct {
		stream  rpcc.Stream
		decoder Decoder
	}
)

func NewEventStream(stream rpcc.Stream, decoder Decoder) runtime.Stream {
	return &EventStream{stream, decoder}
}

func (e *EventStream) Close() error {
	return e.stream.Close()
}

func (e *EventStream) Read(ctx context.Context) <-chan runtime.Message {
	ch := make(chan runtime.Message)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-e.stream.Ready():
				val, err := e.decoder(ctx, e.stream)

				if err != nil {
					ch <- runtime.NewErrorMessage(err)

					return
				}

				if val != nil && val != runtime.None {
					ch <- runtime.NewValueMessage(val)
				}
			}
		}
	}()

	return ch
}
