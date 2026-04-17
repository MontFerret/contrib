package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/mafredri/cdp/rpcc"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	events2 "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/ferret/v2/pkg/runtime"

	. "github.com/smartystreets/goconvey/convey"
)

type (
	TestStream struct {
		ready   chan struct{}
		message chan runtime.Message
		mock.Mock
	}
)

func NewTestStream() *TestStream {
	return NewBufferedTestStream(0)
}

func NewBufferedTestStream(buffer int) *TestStream {
	es := new(TestStream)
	es.ready = make(chan struct{}, buffer)
	es.message = make(chan runtime.Message, buffer)
	return es
}

func (ts *TestStream) Ready() <-chan struct{} {
	return ts.ready
}

func (ts *TestStream) RecvMsg(m any) error {
	return nil
}

func (ts *TestStream) Close() error {
	ts.Called()
	close(ts.message)
	close(ts.ready)
	return nil
}

func (ts *TestStream) Emit(val runtime.Value) {
	ts.ready <- struct{}{}
	ts.message <- runtime.NewValueMessage(val)
}

func (ts *TestStream) EmitError(err error) {
	ts.ready <- struct{}{}
	ts.message <- runtime.NewErrorMessage(err)
}

func (ts *TestStream) Recv() (runtime.Value, error) {
	msg := <-ts.message

	return msg.Value(), msg.Err()
}

func TestStreamReader(t *testing.T) {
	Convey("StreamReader", t, func() {
		Convey("Should read data from Stream", func() {
			ctx, cancel := context.WithCancel(context.Background())

			stream := NewTestStream()
			stream.On("Close").Maybe().Return(nil)

			go func() {
				stream.Emit(runtime.NewString("foo"))
				stream.Emit(runtime.NewString("bar"))
				stream.Emit(runtime.NewString("baz"))
			}()

			data := make([]string, 0, 3)

			es := events2.NewEventStream(stream, func(_ context.Context, stream rpcc.Stream) (runtime.Value, error) {
				return stream.(*TestStream).Recv()
			})

			// Cancel once the expected number of messages has been consumed so
			// the reader goroutine exits and closes the downstream channel.
			for evt := range es.Read(ctx) {
				So(evt.Err(), ShouldBeNil)
				So(evt.Value(), ShouldNotBeNil)

				data = append(data, evt.Value().String())

				if len(data) == 3 {
					cancel()
				}
			}

			So(data, ShouldResemble, []string{"foo", "bar", "baz"})

			stream.AssertExpectations(t)

			So(es.Close(), ShouldBeNil)
		})

		Convey("Should handle error but do not close Stream", func() {
			ctx := context.Background()

			stream := NewTestStream()
			stream.On("Close").Maybe().Return(nil)

			go func() {
				stream.EmitError(errors.New("foo"))
			}()

			reader := events2.NewEventStream(stream, func(_ context.Context, stream rpcc.Stream) (runtime.Value, error) {
				return stream.(*TestStream).Recv()
			})

			ch := reader.Read(ctx)
			evt := <-ch
			So(evt.Err(), ShouldNotBeNil)

			time.Sleep(time.Duration(100) * time.Millisecond)

			stream.AssertExpectations(t)
		})

		Convey("Should not close Stream when Context is cancelled", func() {
			stream := NewTestStream()
			stream.On("Close").Maybe().Return(nil)

			reader := events2.NewEventStream(stream, func(_ context.Context, stream rpcc.Stream) (runtime.Value, error) {
				return runtime.EmptyArray(), nil
			})

			ctx, cancel := context.WithCancel(context.Background())

			_ = reader.Read(ctx)

			time.Sleep(time.Duration(100) * time.Millisecond)

			cancel()

			time.Sleep(time.Duration(100) * time.Millisecond)

			stream.AssertExpectations(t)
		})
	})
}
