package dom

import (
	"context"
	"errors"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeDOMBindingRuntime struct {
	removeErr   error
	removeCalls int
}

func (r *fakeDOMBindingRuntime) AddBinding(_ context.Context, _ *cdpruntime.AddBindingArgs) error {
	return nil
}

func (r *fakeDOMBindingRuntime) RemoveBinding(_ context.Context, _ *cdpruntime.RemoveBindingArgs) error {
	r.removeCalls++

	return r.removeErr
}

func (r *fakeDOMBindingRuntime) BindingCalled(_ context.Context) (cdpruntime.BindingCalledClient, error) {
	return nil, nil
}

func TestDOMEventStreamClose(t *testing.T) {
	Convey("domEventStream.Close", t, func() {
		Convey("Should cleanup once and ignore missing target/context errors", func() {
			stream := newFakeBindingCalledStream(0)
			runtimeAPI := &fakeDOMBindingRuntime{
				removeErr: errors.New("Cannot find context with specified id"),
			}
			cleanupCalls := 0

			reader := &domEventStream{
				stream: stream,
				cleanup: func() error {
					cleanupCalls++

					return closeDOMEventResources(
						runtimeAPI,
						"match",
						func(_ context.Context, _ string) error {
							return errors.New("execution context was destroyed")
						},
						stream,
						nil,
					)
				},
			}

			So(reader.Close(), ShouldBeNil)
			So(reader.Close(), ShouldBeNil)
			So(cleanupCalls, ShouldEqual, 1)
			So(runtimeAPI.removeCalls, ShouldEqual, 1)
			So(stream.CloseCalls(), ShouldEqual, 1)
		})
	})
}
