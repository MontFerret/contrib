package drivers_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/contrib/modules/html/drivers"
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
				So(string(out), ShouldEqual, `{"domain":"montferret.dev","expires":"0001-01-01T00:00:00Z","http_only":true,"max_age":320,"name":"test_cookie","path":"/","same_site":"Lax","secure":true,"Value":"test_value"}`)
			})
		})
	})
}
