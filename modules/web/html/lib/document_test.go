package lib

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const maxNavigationTimeoutMillis = 9_223_372_036_854

func TestNewPageLoadParamsTimeout(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value runtime.Value
		want  time.Duration
	}{
		{name: "omitted", want: 60 * time.Second},
		{name: "five seconds", value: runtime.Int(5000), want: 5 * time.Second},
		{name: "one millisecond", value: runtime.Int(1), want: time.Millisecond},
		{name: "maximum", value: runtime.Int(maxNavigationTimeoutMillis), want: maxNavigationTimeoutMillis * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options := runtime.NewObject()
			if tc.value != nil {
				if err := options.Set(t.Context(), runtime.String("timeout"), tc.value); err != nil {
					t.Fatal(err)
				}
			}

			params, err := newPageLoadParams(t.Context(), "https://example.com", options)
			if assert.NoError(t, err) {
				assert.Equal(t, tc.want, params.Timeout)
			}
		})
	}
}

func TestOpenRejectsInvalidTimeout(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   runtime.Value
		wantErr error
	}{
		{name: "zero", value: runtime.Int(0), wantErr: runtime.ErrInvalidArgument},
		{name: "negative", value: runtime.Int(-1), wantErr: runtime.ErrInvalidArgument},
		{name: "overflow", value: runtime.Int(maxNavigationTimeoutMillis + 1), wantErr: runtime.ErrInvalidArgument},
		{name: "string", value: runtime.String("5000"), wantErr: runtime.ErrInvalidArgumentType},
		{name: "float", value: runtime.Float(5000), wantErr: runtime.ErrInvalidArgumentType},
		{name: "boolean", value: runtime.True, wantErr: runtime.ErrInvalidArgumentType},
		{name: "duration", value: runtime.NewDuration(5 * time.Second), wantErr: runtime.ErrInvalidArgumentType},
		{name: "none", value: runtime.None, wantErr: runtime.ErrInvalidArgumentType},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, drv := newRecordingDriver(t)
			options := runtime.NewObjectWith(map[string]runtime.Value{"timeout": tc.value})

			_, err := Open(ctx, runtime.String("https://example.com"), options)
			if assert.ErrorIs(t, err, tc.wantErr) {
				assert.ErrorContains(t, err, "timeout")
			}

			assert.Nil(t, drv.openContext)
		})
	}
}

func TestOpenNavigationDeadline(t *testing.T) {
	for _, tc := range []struct {
		name          string
		options       runtime.Value
		parentTimeout time.Duration
		wantTimeout   time.Duration
	}{
		{name: "no options", wantTimeout: 60 * time.Second},
		{
			name:        "driver option without timeout",
			options:     runtime.NewObjectWith(map[string]runtime.Value{"driver": runtime.String("recording")}),
			wantTimeout: 60 * time.Second,
		},
		{name: "driver shorthand", options: runtime.String("recording"), wantTimeout: 60 * time.Second},
		{
			name:        "explicit five seconds",
			options:     runtime.NewObjectWith(map[string]runtime.Value{"timeout": runtime.Int(5000)}),
			wantTimeout: 5 * time.Second,
		},
		{
			name:        "explicit one millisecond",
			options:     runtime.NewObjectWith(map[string]runtime.Value{"timeout": runtime.Int(1)}),
			wantTimeout: time.Millisecond,
		},
		{name: "earlier caller deadline", parentTimeout: 2 * time.Second, wantTimeout: 2 * time.Second},
		{
			name:          "later caller deadline",
			options:       runtime.NewObjectWith(map[string]runtime.Value{"timeout": runtime.Int(5000)}),
			parentTimeout: 10 * time.Second,
			wantTimeout:   5 * time.Second,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, drv := newRecordingDriver(t)
				args := []runtime.Value{runtime.String("https://example.com")}

				if tc.options != nil {
					args = append(args, tc.options)
				}

				start := time.Now()
				if tc.parentTimeout != 0 {
					parent, cancel := context.WithDeadline(ctx, start.Add(tc.parentTimeout))
					defer cancel()
					ctx = parent
				}

				_, err := Open(ctx, args...)
				assert.NoError(t, err)

				if drv.openContext == nil {
					t.Fatal("driver was not called")
				}

				deadline, ok := drv.openContext.Deadline()
				if assert.True(t, ok, "driver context has no deadline") {
					assert.Equal(t, tc.wantTimeout, deadline.Sub(start))
				}

				assert.ErrorIs(t, drv.openContext.Err(), context.Canceled)
			})
		})
	}
}

func TestOpenPropagatesParentCancellation(t *testing.T) {
	ctx, drv := newRecordingDriver(t)
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	_, err := Open(ctx, runtime.String("https://example.com"))
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotNil(t, drv.openContext)
}

func TestNewPageLoadParamsInitScript(t *testing.T) {
	ctx := context.Background()
	url := runtime.NewString("https://example.com")

	t.Run("defaults timing", func(t *testing.T) {
		params, err := newPageLoadParams(ctx, url, runtime.NewObjectWith(map[string]runtime.Value{
			"initScript": runtime.NewObjectWith(map[string]runtime.Value{
				"source": runtime.NewString("window.ready = true"),
			}),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.InitScript == nil || params.InitScript.Timing != drivers.InitScriptAfterNavigation {
			t.Fatalf("unexpected initScript: %#v", params.InitScript)
		}
	})

	for _, tt := range []struct {
		script map[string]runtime.Value
		name   string
	}{
		{name: "missing source", script: map[string]runtime.Value{"timing": runtime.NewString("beforeDocument")}},
		{name: "blank source", script: map[string]runtime.Value{"source": runtime.NewString("  ")}},
		{name: "invalid timing", script: map[string]runtime.Value{"source": runtime.NewString("true"), "timing": runtime.NewString("beforeLoad")}},
		{name: "unknown nested field", script: map[string]runtime.Value{"source": runtime.NewString("true"), "evaluateArgs": runtime.NewArray(0)}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newPageLoadParams(ctx, url, runtime.NewObjectWith(map[string]runtime.Value{
				"initScript": runtime.NewObjectWith(tt.script),
			}))
			if err == nil {
				t.Fatal("expected decoding or validation error")
			}
		})
	}

	t.Run("evaluateArgs alias is rejected", func(t *testing.T) {
		_, err := newPageLoadParams(ctx, url, runtime.NewObjectWith(map[string]runtime.Value{
			"evaluateArgs": runtime.NewArray(0),
		}))
		if err == nil {
			t.Fatal("expected unknown field error")
		}
		if errors.Is(err, runtime.ErrNotSupported) {
			t.Fatalf("expected strict decoding error, got %v", err)
		}
	})
}

func TestDocument(t *testing.T) {
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
						"theme": {
							Name:  "theme",
							Value: "dark",
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
				Timeout: 5 * time.Second,
			}
			out, err := newPageLoadParams(context.Background(), "https://example.com", runtime.NewObjectWith(map[string]runtime.Value{
				"driver":      runtime.NewString(expected.Driver),
				"userAgent":   runtime.NewString(expected.UserAgent),
				"keepCookies": runtime.True,
				"timeout":     runtime.Int(5000),
				"charset":     runtime.NewString(expected.Charset),
				"headers": runtime.NewObjectWith(map[string]runtime.Value{
					"Accept": runtime.String("text/html"),
				}),
				"viewport": runtime.NewObjectWith(map[string]runtime.Value{
					"height": runtime.Int(1024), "width": runtime.Int(768), "scaleFactor": runtime.Float(1),
					"mobile": runtime.True, "landscape": runtime.True,
				}),
				"cookies": runtime.NewArrayWith(
					runtime.NewObjectWith(map[string]runtime.Value{
						"name":   runtime.NewString("session"),
						"value":  runtime.NewString("abc123"),
						"domain": runtime.NewString("example.com"),
					}),
					runtime.NewObjectWith(map[string]runtime.Value{
						"name":  runtime.NewString("theme"),
						"value": runtime.NewString("dark"),
					}),
				),
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
				t.Fatalf("expected timeout %v, got %v", expected.Timeout, out.Timeout)
			}

			if !reflect.DeepEqual(out.Headers, expected.Headers) {
				t.Fatalf("headers = %#v, want %#v", out.Headers, expected.Headers)
			}

			if !reflect.DeepEqual(out.Viewport, expected.Viewport) {
				t.Fatalf("viewport = %#v, want %#v", out.Viewport, expected.Viewport)
			}

			if out.Cookies == nil {
				t.Fatalf("expected cookies to be set, got nil")
			}

			if len(out.Cookies.Data) != len(expected.Cookies.Data) {
				t.Fatalf("expected %d cookies, got %d", len(expected.Cookies.Data), len(out.Cookies.Data))
			}

			if out.Cookies.Data["session"].Value != "abc123" {
				t.Fatalf("expected session cookie to be parsed, got %#v", out.Cookies.Data["session"])
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
