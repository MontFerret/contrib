package drivers_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestHTTPCookie(t *testing.T) {
	Convey("HTTPCookie", t, func() {
		Convey(".MarshalJSON", func() {
			Convey("Should serialize cookie Data", func() {
				cookie := &drivers.HTTPCookie{}

				cookie.Name = "test_cookie"
				cookie.Value = "test_value"
				cookie.Domain = "montferret.dev"
				cookie.HTTPOnly = true
				cookie.MaxAge = 320
				cookie.Path = "/"
				cookie.SameSite = drivers.SameSiteLaxMode
				cookie.Secure = true

				out, err := cookie.MarshalJSON()

				So(err, ShouldBeNil)
				So(string(out), ShouldEqual, `{"domain":"montferret.dev","expires":"0001-01-01T00:00:00Z","httpOnly":true,"maxAge":320,"name":"test_cookie","path":"/","sameSite":"Lax","secure":true,"value":"test_value"}`)
			})
		})

		Convey(".Get", func() {
			Convey("Should expose lowercase runtime fields", func() {
				cookie := drivers.HTTPCookie{
					Name:     "test_cookie",
					Value:    "test_value",
					HTTPOnly: true,
				}

				value, err := cookie.Get(context.Background(), runtime.NewString("value"))
				So(err, ShouldBeNil)
				So(value, ShouldEqual, runtime.NewString("test_value"))

				httpOnly, err := cookie.Get(context.Background(), runtime.NewString("httpOnly"))
				So(err, ShouldBeNil)
				So(httpOnly, ShouldEqual, runtime.True)

				legacy, err := cookie.Get(context.Background(), runtime.NewString("Value"))
				So(err, ShouldBeNil)
				So(legacy, ShouldEqual, runtime.None)
			})
		})
	})
}
