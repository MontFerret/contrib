package dom

import (
	"context"
	"sync"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
)

type (
	domBindingRuntime interface {
		AddBinding(ctx context.Context, args *cdpruntime.AddBindingArgs) error
		RemoveBinding(ctx context.Context, args *cdpruntime.RemoveBindingArgs) error
		BindingCalled(ctx context.Context) (cdpruntime.BindingCalledClient, error)
	}

	domEventSubscription func(ctx context.Context, bindingName string) error

	domEventStream struct {
		bindingName string
		contextID   cdpruntime.ExecutionContextID
		stream      cdpruntime.BindingCalledClient
		cleanup     func() error
		closeOnce   sync.Once
		closeErr    error
	}
)

func subscribeDOMEvents(
	ctx context.Context,
	api domBindingRuntime,
	contextID cdpruntime.ExecutionContextID,
	attach domEventSubscription,
	detach domEventSubscription,
) (runtime.Stream, error) {
	bindingName, err := newDOMBindingName()

	if err != nil {
		return nil, err
	}

	if err := api.AddBinding(ctx, cdpruntime.NewAddBindingArgs(bindingName)); err != nil {
		return nil, err
	}

	stream, err := api.BindingCalled(ctx)

	if err != nil {
		return nil, closeDOMEventResources(api, bindingName, nil, nil, err)
	}

	if err := attach(ctx, bindingName); err != nil {
		return nil, closeDOMEventResources(api, bindingName, nil, stream, err)
	}

	return &domEventStream{
		bindingName: bindingName,
		contextID:   contextID,
		stream:      stream,
		cleanup: func() error {
			return closeDOMEventResources(api, bindingName, detach, stream, nil)
		},
	}, nil
}

func (s *domEventStream) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.cleanup()
	})

	return s.closeErr
}

func (s *domEventStream) Read(ctx context.Context) <-chan runtime.Message {
	ch := make(chan runtime.Message)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stream.Ready():
				reply, err := s.stream.Recv()

				if err != nil {
					ch <- runtime.NewErrorMessage(err)

					return
				}

				if reply.Name != s.bindingName || reply.ExecutionContextID != s.contextID {
					continue
				}

				val, err := decodeDOMEventPayload(reply.Payload)

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
