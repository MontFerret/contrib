package dom

import (
	"context"
	"io"
	"sync"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/ferret/v2/pkg/runtime"

	. "github.com/smartystreets/goconvey/convey"
)

type (
	bindingCalledResult struct {
		reply *cdpruntime.BindingCalledReply
		err   error
	}

	fakeBindingCalledStream struct {
		closeErr   error
		ready      chan struct{}
		results    chan bindingCalledResult
		closeCalls int
		mu         sync.Mutex
		closed     bool
	}
)

func newFakeBindingCalledStream(buffer int) *fakeBindingCalledStream {
	return &fakeBindingCalledStream{
		ready:   make(chan struct{}, buffer),
		results: make(chan bindingCalledResult, buffer),
	}
}

func (s *fakeBindingCalledStream) Ready() <-chan struct{} {
	return s.ready
}

func (s *fakeBindingCalledStream) RecvMsg(_ any) error {
	return nil
}

func (s *fakeBindingCalledStream) Recv() (*cdpruntime.BindingCalledReply, error) {
	result, ok := <-s.results

	if !ok {
		return nil, io.EOF
	}

	return result.reply, result.err
}

func (s *fakeBindingCalledStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closeCalls++

	if s.closed {
		return s.closeErr
	}

	s.closed = true
	close(s.results)
	close(s.ready)

	return s.closeErr
}

func (s *fakeBindingCalledStream) Emit(reply *cdpruntime.BindingCalledReply) {
	s.results <- bindingCalledResult{reply: reply}
	s.ready <- struct{}{}
}

func (s *fakeBindingCalledStream) CloseCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closeCalls
}

func TestDOMEventStreamRead(t *testing.T) {
	Convey("domEventStream.Read", t, func() {
		Convey("Should ignore events from other bindings", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			stream := newFakeBindingCalledStream(2)
			reader := &domEventStream{
				bindingName: "match",
				contextID:   42,
				stream:      stream,
				cleanup: func() error {
					return nil
				},
			}

			ch := reader.Read(ctx)

			stream.Emit(&cdpruntime.BindingCalledReply{
				Name:               "other",
				ExecutionContextID: 42,
				Payload:            `{"type":"ignored"}`,
			})
			stream.Emit(&cdpruntime.BindingCalledReply{
				Name:               "match",
				ExecutionContextID: 42,
				Payload:            `{"type":"ferret:element"}`,
			})

			msg := <-ch

			So(msg.Err(), ShouldBeNil)

			value, err := runtime.CastMap(msg.Value())
			So(err, ShouldBeNil)

			eventType, err := value.Get(ctx, runtime.NewString("type"))
			So(err, ShouldBeNil)
			So(eventType, ShouldEqual, runtime.NewString("ferret:element"))
		})

		Convey("Should ignore events from other execution contexts", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			stream := newFakeBindingCalledStream(2)
			reader := &domEventStream{
				bindingName: "match",
				contextID:   42,
				stream:      stream,
				cleanup: func() error {
					return nil
				},
			}

			ch := reader.Read(ctx)

			stream.Emit(&cdpruntime.BindingCalledReply{
				Name:               "match",
				ExecutionContextID: 7,
				Payload:            `{"type":"ignored"}`,
			})
			stream.Emit(&cdpruntime.BindingCalledReply{
				Name:               "match",
				ExecutionContextID: 42,
				Payload:            `{"type":"ferret:document"}`,
			})

			msg := <-ch

			So(msg.Err(), ShouldBeNil)

			value, err := runtime.CastMap(msg.Value())
			So(err, ShouldBeNil)

			eventType, err := value.Get(ctx, runtime.NewString("type"))
			So(err, ShouldBeNil)
			So(eventType, ShouldEqual, runtime.NewString("ferret:document"))
		})

		Convey("Should decode binding payload into runtime values", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			stream := newFakeBindingCalledStream(1)
			reader := &domEventStream{
				bindingName: "match",
				contextID:   42,
				stream:      stream,
				cleanup: func() error {
					return nil
				},
			}

			ch := reader.Read(ctx)

			stream.Emit(&cdpruntime.BindingCalledReply{
				Name:               "match",
				ExecutionContextID: 42,
				Payload:            `{"type":"ferret:sequence","detail":{"index":2}}`,
			})

			msg := <-ch

			So(msg.Err(), ShouldBeNil)

			value, err := runtime.CastMap(msg.Value())
			So(err, ShouldBeNil)

			eventType, err := value.Get(ctx, runtime.NewString("type"))
			So(err, ShouldBeNil)
			So(eventType, ShouldEqual, runtime.NewString("ferret:sequence"))

			detailValue, err := value.Get(ctx, runtime.NewString("detail"))
			So(err, ShouldBeNil)

			detail, err := runtime.CastMap(detailValue)
			So(err, ShouldBeNil)

			index, err := detail.Get(ctx, runtime.NewString("index"))
			So(err, ShouldBeNil)
			So(index, ShouldEqual, runtime.NewInt(2))
		})
	})
}
