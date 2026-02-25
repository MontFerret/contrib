package html

import (
	"testing"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDocument(t *testing.T) {
	defaultTimeout := drivers.DefaultPageLoadTimeout * time.Millisecond

	t.Run("newPageLoadParams", func(t *testing.T) {
		t.Run("all params", func(t *testing.T) {
			expected := PageLoadParams{
				Params: drivers.Params{
					URL:         "https://example.com",
					UserAgent:   "test-agent",
					KeepCookies: true,
					Cookies: drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{
						"session": {
							Name:   "session",
							Value:  "abc123",
							Domain: "example.com",
						},
					}),
					Headers: drivers.NewHTTPHeadersWith(map[string][]string{
						"Accept": {"text/html"},
					}),
					Viewport: &drivers.Viewport{
						Height:      1024,
						Width:       768,
						ScaleFactor: 1.0,
						Mobile:      true,
						Landscape:   true,
					},
					Charset: "utf-8",
					Ignore: &drivers.Ignore{
						Resources: []drivers.ResourceFilter{{URL: "https://example.com/ads", Type: "image"}},
						StatusCodes: []drivers.StatusCodeFilter{
							{
								URL:  "https://example.com/api",
								Code: 500,
							},
						},
					},
				},
				Driver:  "cdp",
				Timeout: 5000,
			}
			out, err := newPageLoadParams("https://example.com", runtime.NewObjectWith(map[string]runtime.Value{
				"driver":      runtime.NewString(expected.Driver),
				"userAgent":   runtime.NewString(expected.UserAgent),
				"keepCookies": runtime.True,
				"timeout":     runtime.NewInt(int(expected.Timeout)),
				"charset":     runtime.NewString(expected.Charset),
				"ignore": runtime.NewObjectWith(map[string]runtime.Value{
					"resources": runtime.NewArrayWith(
						runtime.NewObjectWith(map[string]runtime.Value{
							"url":  runtime.NewString(expected.Ignore.Resources[0].URL),
							"type": runtime.NewString(expected.Ignore.Resources[0].Type),
						}),
					),
					"statusCodes": runtime.NewArrayWith(
						runtime.NewObjectWith(map[string]runtime.Value{
							"url":  runtime.NewString(expected.Ignore.StatusCodes[0].URL),
							"code": runtime.NewInt(expected.Ignore.StatusCodes[0].Code),
						}),
					),
				}),
			}))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if out.URL != expected.URL {
				t.Fatalf("expected URL %s, got %s", expected.URL, out.URL)
			}

			if out.Timeout != expected.Timeout {
				t.Fatalf("expected timeout %v, got %v", defaultTimeout, out.Timeout)
			}

			if out.Ignore == nil {
				t.Fatalf("expected ignore to be set, got nil")
			}

			if out.Ignore.Resources == nil {
				t.Fatalf("expected ignore resources to be set, got nil")
			}

			if len(out.Ignore.Resources) != len(expected.Ignore.Resources) {
				t.Fatalf("expected %d ignore resources, got %d", len(expected.Ignore.Resources), len(out.Ignore.Resources))
			}
		})
	})
}
