package drivers_test

import (
	"testing"

	"github.com/goccy/go-json"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

func TestHTTPHeaders(t *testing.T) {
	Convey("HTTPHeaders", t, func() {
		Convey(".MarshalJSON", func() {
			Convey("Should serialize header Data", func() {
				headers := drivers.NewHTTPHeadersWith(map[string][]string{
					"Content-Encoding": []string{"gzip"},
					"Content-Type":     []string{"text/html", "charset=utf-8"},
				})

				out, err := headers.MarshalJSON()

				So(err, ShouldBeNil)
				So(string(out), ShouldEqual, `{"Content-Encoding":"gzip","Content-Type":"text/html, charset=utf-8"}`)
			})

			Convey("Should set proper Data", func() {
				headers := drivers.NewHTTPHeaders()

				headers.Set("Authorization", `["Basic e40b7d5eff464a4fb51efed2d1a19a24"]`)

				_, err := json.Marshal(headers)

				So(err, ShouldBeNil)
			})
		})
	})
}
