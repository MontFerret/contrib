package memory_test

import (
	stdhttp "net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestNewOptions(t *testing.T) {
	Convey("Should create driver options with initial values", t, func() {
		opts := memory.NewOptions([]memory.Option{})
		So(opts.Options, ShouldNotBeNil)
		So(opts.Name, ShouldEqual, memory.DriverName)
		//So(opts.Backoff, ShouldEqual, pester.ExponentialBackoff)
		So(opts.Concurrency, ShouldEqual, memory.DefaultConcurrency)
		So(opts.MaxRetries, ShouldEqual, memory.DefaultMaxRetries)
		So(opts.HTTPCodesFilter, ShouldHaveLength, 0)
	})

	Convey("Should use setters to set values", t, func() {
		expectedName := memory.DriverName + "2"
		expectedUA := "Mozilla"
		expectedProxy := "https://proxy.com"
		expectedMaxRetries := 2
		expectedConcurrency := 10
		expectedTransport := &stdhttp.Transport{}
		expectedTimeout := time.Second * 5

		opts := memory.NewOptions([]memory.Option{
			memory.WithCustomName(expectedName),
			memory.WithUserAgent(expectedUA),
			memory.WithProxy(expectedProxy),
			memory.WithCookie(drivers.HTTPCookie{
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
			memory.WithCookies([]drivers.HTTPCookie{
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
			memory.WithHeader("Authorization", []string{"Bearer dfsd7f98sd9fsd9fsd"}),
			memory.WithHeaders(drivers.NewHTTPHeadersWith(map[string][]string{
				"x-correlation-id": {"232483833833839"},
			})),
			memory.WithDefaultBackoff(),
			memory.WithMaxRetries(expectedMaxRetries),
			memory.WithConcurrency(expectedConcurrency),
			memory.WithAllowedHTTPCode(401),
			memory.WithAllowedHTTPCodes([]int{403, 404}),
			memory.WithCustomTransport(expectedTransport),
			memory.WithTimeout(time.Second * 5),
		})
		So(opts.Options, ShouldNotBeNil)
		So(opts.Name, ShouldEqual, expectedName)
		So(opts.UserAgent, ShouldEqual, expectedUA)
		So(opts.Proxy, ShouldEqual, expectedProxy)
		So(len(opts.Cookies.Data), ShouldEqual, 2)
		So(len(opts.Headers.Data), ShouldEqual, 2)
		//So(opts.Backoff, ShouldEqual, pester.DefaultBackoff)
		So(opts.MaxRetries, ShouldEqual, expectedMaxRetries)
		So(opts.Concurrency, ShouldEqual, expectedConcurrency)
		So(opts.HTTPCodesFilter, ShouldHaveLength, 3)
		So(opts.HTTPTransport, ShouldEqual, expectedTransport)
		So(opts.Timeout, ShouldEqual, expectedTimeout)
	})
}
