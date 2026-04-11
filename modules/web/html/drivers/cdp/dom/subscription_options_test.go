package dom

import (
	"context"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseDOMEventOptions(t *testing.T) {
	Convey("parseDOMEventOptions", t, func() {
		ctx := context.Background()

		Convey("Should extract listener options and preserve them in template config", func() {
			listener := runtime.NewObjectWith(map[string]runtime.Value{
				"once":    runtime.True,
				"capture": runtime.True,
			})

			options, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionListener: listener,
				domEventOptionMaxDepth: runtime.NewInt(2),
			}))

			So(err, ShouldBeNil)
			So(options.Listener, ShouldEqual, listener)
			So(options.MaxDepth, ShouldEqual, runtime.NewInt(2))

			config := buildDOMEventTemplateOptions(options)

			listenerValue, err := config.Get(ctx, runtime.NewString(domEventOptionListener))
			So(err, ShouldBeNil)

			actualListener, err := runtime.CastMap(listenerValue)
			So(err, ShouldBeNil)

			once, err := actualListener.Get(ctx, runtime.NewString("once"))
			So(err, ShouldBeNil)
			So(once, ShouldEqual, runtime.True)

			capture, err := actualListener.Get(ctx, runtime.NewString("capture"))
			So(err, ShouldBeNil)
			So(capture, ShouldEqual, runtime.True)
		})

		Convey("Should reject unknown top-level options", func() {
			_, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				"once": runtime.True,
			}))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unknown DOM event option")
		})

		Convey("Should validate listener option type", func() {
			_, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionListener: runtime.NewString("invalid"),
			}))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, domEventOptionListener)
		})

		Convey("Should reject delegate and targetSelector together", func() {
			_, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionDelegate:       runtime.NewString(".item"),
				domEventOptionTargetSelector: runtime.NewString(".child"),
			}))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "delegate and targetSelector cannot be used together")
		})

		Convey("Should validate and deduplicate props", func() {
			options, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionProps: runtime.NewArrayWith(
					runtime.NewString("detail"),
					runtime.NewString("type"),
					runtime.NewString("detail"),
				),
			}))

			So(err, ShouldBeNil)
			So(options.HasProps, ShouldBeTrue)

			length, err := options.Props.Length(ctx)
			So(err, ShouldBeNil)
			So(length, ShouldEqual, runtime.NewInt(2))

			first, err := options.Props.At(ctx, runtime.NewInt(0))
			So(err, ShouldBeNil)
			So(first, ShouldEqual, runtime.NewString("detail"))

			second, err := options.Props.At(ctx, runtime.NewInt(1))
			So(err, ShouldBeNil)
			So(second, ShouldEqual, runtime.NewString("type"))
		})

		Convey("Should reject non-string props", func() {
			_, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionProps: runtime.NewArrayWith(runtime.NewInt(1)),
			}))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, domEventOptionProps)
		})

		Convey("Should enforce maxDepth bounds", func() {
			_, err := parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionMaxDepth: runtime.NewInt(0),
			}))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, domEventOptionMaxDepth)

			_, err = parseDOMEventOptions(ctx, runtime.NewObjectWith(map[string]runtime.Value{
				domEventOptionMaxDepth: runtime.NewInt(9),
			}))

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, domEventOptionMaxDepth)
		})
	})
}
