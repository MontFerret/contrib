package cdp_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
)

func TestNewOptions(t *testing.T) {
	Convey("Should create driver options with initial values", t, func() {
		opts := cdp.NewOptions([]cdp.Option{})
		So(opts.Options, ShouldNotBeNil)
		So(opts.Name, ShouldEqual, cdp.DriverName)
		So(opts.Address, ShouldEqual, cdp.DefaultAddress)
	})

	Convey("Should use setters to set values", t, func() {
		expectedName := cdp.DriverName + "2"
		expectedAddress := "0.0.0.0:9222"
		expectedUA := "Mozilla"
		expectedProxy := "https://proxy.com"

		opts := cdp.NewOptions([]cdp.Option{
			cdp.WithCustomName(expectedName),
			cdp.WithAddress(expectedAddress),
			cdp.WithUserAgent(expectedUA),
			cdp.WithProxy(expectedProxy),
			cdp.WithKeepCookies(),
			cdp.WithCookie(drivers.HTTPCookie{
				Name:     "Session",
				Value:    "fsdfsdfs",
				Path:     "dfsdfsd",
				Domain:   "sfdsfs",
				Expires:  time.Time{},
				MaxAge:   0,
				Secure:   false,
				HTTPOnly: false,
				SameSite: 0,
			}),
			cdp.WithCookies([]drivers.HTTPCookie{
				{
					Name:     "Use",
					Value:    "Foos",
					Path:     "",
					Domain:   "",
					Expires:  time.Time{},
					MaxAge:   0,
					Secure:   false,
					HTTPOnly: false,
					SameSite: 0,
				},
			}),
			cdp.WithHeader("Authorization", []string{"Bearer dfsd7f98sd9fsd9fsd"}),
			cdp.WithHeaders(drivers.NewHTTPHeadersWith(map[string][]string{
				"x-correlation-id": {"232483833833839"},
			}).Data),
		})
		So(opts.Options, ShouldNotBeNil)
		So(opts.Name, ShouldEqual, expectedName)
		So(opts.Address, ShouldEqual, expectedAddress)
		So(opts.UserAgent, ShouldEqual, expectedUA)
		So(opts.Proxy, ShouldEqual, expectedProxy)
		So(opts.KeepCookies, ShouldBeTrue)
		So(len(opts.Cookies.Data), ShouldEqual, 2)
		So(len(opts.Headers.Data), ShouldEqual, 2)
	})

	Convey("Should normalize loopback address to localhost", t, func() {
		opts := cdp.NewOptions([]cdp.Option{
			cdp.WithAddress("http://127.0.0.1:9222"),
		})

		So(opts.Address, ShouldEqual, "http://localhost:9222")
	})
}
