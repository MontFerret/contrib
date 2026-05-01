package styleutil_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/internal/styleutil"
	"github.com/MontFerret/ferret/v2/pkg/runtime"

	. "github.com/smartystreets/goconvey/convey"
)

type style struct {
	value runtime.Value
	raw   string
	name  runtime.String
}

func TestDeserialize(t *testing.T) {
	Convey("Deserialize", t, func() {
		ctx := context.Background()
		styles := []style{
			{raw: "min-height: 1.15", name: "min-height", value: runtime.NewFloat(1.15)},
			{raw: "background-color: #4A154B", name: "background-color", value: runtime.NewString("#4A154B")},
			{raw: "font-size:26pt", name: "font-size", value: runtime.NewString("26pt")},
			{raw: "page-break-after:avoid", name: "page-break-after", value: runtime.NewString("avoid")},
			{raw: `font-family: Arial,"Helvetica Neue",Helvetica,sans-serif`, name: "font-family", value: runtime.NewString(`Arial,"Helvetica Neue",Helvetica,sans-serif`)},
			{raw: "color: black", name: "color", value: runtime.NewString("black")},
			{raw: "display: inline-block", name: "display", value: runtime.NewString("inline-block")},
			{raw: "min-width: 50", name: "min-width", value: runtime.NewFloat(50)},
		}

		Convey("Should parse a single style", func() {
			for _, s := range styles {
				out, err := styleutil.Deserialize(ctx, runtime.NewString(s.raw))

				So(err, ShouldBeNil)
				So(out, ShouldNotBeNil)

				exists, err := out.ContainsKey(ctx, s.name)
				So(err, ShouldBeNil)
				So(bool(exists), ShouldBeTrue)

				value, err := out.Get(ctx, s.name)
				So(err, ShouldBeNil)

				So(runtime.CompareValues(value, s.value), ShouldEqual, 0)
			}
		})

		Convey("Should parse multiple styles", func() {
			var buff bytes.Buffer

			for _, s := range styles {
				buff.WriteString(s.raw)
				buff.WriteString("; ")
			}

			out, err := styleutil.Deserialize(ctx, runtime.NewString(buff.String()))

			So(err, ShouldBeNil)
			So(out, ShouldNotBeNil)

			length, err := out.Length(ctx)
			So(err, ShouldBeNil)
			So(length, ShouldEqual, runtime.NewInt(len(styles)))

			for _, s := range styles {
				exists, err := out.ContainsKey(ctx, s.name)
				So(err, ShouldBeNil)
				So(bool(exists), ShouldBeTrue)

				value, err := out.Get(ctx, s.name)
				So(err, ShouldBeNil)

				So(runtime.CompareValues(value, s.value), ShouldEqual, 0)
			}
		})
	})
}
