package drivers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/goccy/go-json"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestHTTPCookies(t *testing.T) {
	Convey("HTTPCookies", t, func() {
		Convey(".MarshalJSON", func() {
			Convey("Should serialize cookies", func() {
				expires := time.Now()
				headers := drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{
					"Session": {
						Name:     "Session",
						Value:    "asdfg",
						Path:     "/",
						Domain:   "www.google.com",
						Expires:  expires,
						MaxAge:   0,
						Secure:   true,
						HTTPOnly: true,
						SameSite: drivers.SameSiteLaxMode,
					},
				})

				out, err := headers.MarshalJSON()

				t, e := expires.MarshalJSON()
				So(e, ShouldBeNil)

				expected := fmt.Sprintf(`{"Session":{"domain":"www.google.com","expires":%s,"httpOnly":true,"maxAge":0,"name":"Session","path":"/","sameSite":"Lax","secure":true,"value":"asdfg"}}`, string(t))

				So(err, ShouldBeNil)
				So(string(out), ShouldEqual, expected)
			})

			Convey("Should set proper Data", func() {
				headers := drivers.NewHTTPCookies()
				cookie := drivers.HTTPCookie{
					Name:     "Authorization",
					Value:    "e40b7d5eff464a4fb51efed2d1a19a24",
					Path:     "/",
					Domain:   "www.google.com",
					Expires:  time.Now(),
					MaxAge:   0,
					Secure:   false,
					HTTPOnly: false,
					SameSite: 0,
				}

				err := headers.Set(
					context.Background(),
					runtime.NewString(cookie.Name),
					cookie,
				)

				So(err, ShouldBeNil)

				_, err = json.Marshal(headers)

				So(err, ShouldBeNil)
			})
		})
	})
}
