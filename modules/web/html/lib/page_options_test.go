package lib

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func TestPageHeaders(t *testing.T) {
	native := drivers.NewHTTPHeadersWith(map[string][]string{"X-Test": {"one", "two"}})

	for _, function := range []string{"Open", "Parse"} {
		for _, tc := range []struct {
			name  string
			value runtime.Value
			want  *drivers.HTTPHeaders
		}{
			{"omitted", nil, nil},
			{"none", runtime.None, nil},
			{"empty", runtime.NewObject(), drivers.NewHTTPHeaders()},
			{"map", runtime.NewObjectWith(map[string]runtime.Value{"x-test": runtime.String("probe")}),
				drivers.NewHTTPHeadersWith(map[string][]string{"X-Test": {"probe"}})},
			{"native", native, native},
			{"wrapped", sdk.NewHostValue(native), native},
		} {
			t.Run(function+"/"+tc.name, func(t *testing.T) {
				options := runtime.NewObject()

				if tc.value != nil {
					if err := options.Set(t.Context(), runtime.String("headers"), tc.value); err != nil {
						t.Fatal(err)
					}
				}

				params, err := callPageOptions(t, function, options)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(params.Headers, tc.want) {
					t.Fatalf("headers = %#v, want %#v", params.Headers, tc.want)
				}
			})
		}
	}
}

func TestPageCookies(t *testing.T) {
	cookie := drivers.HTTPCookie{Name: "session", Value: "probe", SameSite: drivers.SameSiteDefaultMode}
	collection := drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{"session": cookie})

	for _, function := range []string{"Open", "Parse"} {
		for _, tc := range []struct {
			name  string
			value runtime.Value
			want  *drivers.HTTPCookies
		}{
			{"omitted", nil, nil},
			{"none", runtime.None, nil},
			{"object", cookieMap("session", "probe"), collection},
			{"list", runtime.NewArrayWith(cookieMap("session", "probe")), collection},
			{"empty list", runtime.NewArray(0), drivers.NewHTTPCookies()},
			{"native cookie", cookie, collection},
			{"native cookie pointer", &cookie, collection},
			{"collection", collection, collection},
			{"wrapped cookie", sdk.NewHostValue(cookie), collection},
			{"wrapped collection", sdk.NewHostValue(collection), collection},
			{"legacy fields", runtime.NewObjectWith(map[string]runtime.Value{
				"Name": runtime.String("session"), "Value": runtime.String("probe"), "http_only": runtime.True,
				"max_age": runtime.Int(60), "same_site": runtime.String("strict"),
			}), drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{
				"session": {Name: "session", Value: "probe", HTTPOnly: true, MaxAge: 60, SameSite: drivers.SameSiteStrictMode},
			})},
		} {
			t.Run(function+"/"+tc.name, func(t *testing.T) {
				options := runtime.NewObject()

				if tc.value != nil {
					if err := options.Set(t.Context(), runtime.String("cookies"), tc.value); err != nil {
						t.Fatal(err)
					}
				}

				params, err := callPageOptions(t, function, options)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(params.Cookies, tc.want) {
					t.Fatalf("cookies = %#v, want %#v", params.Cookies, tc.want)
				}
			})
		}
	}
}

func TestPageOptionsRejectInvalidValuesBeforeDriver(t *testing.T) {
	for _, function := range []string{"Open", "Parse"} {
		for _, tc := range []struct {
			name    string
			key     string
			value   runtime.Value
			wantErr error
		}{
			{"header container", "headers", runtime.String("secret"), runtime.ErrInvalidArgumentType},
			{"header value", "headers", runtime.NewObjectWith(map[string]runtime.Value{"X-Test": runtime.Int(1)}), runtime.ErrInvalidArgumentType},
			{"header list", "headers", runtime.NewObjectWith(map[string]runtime.Value{"X-Test": runtime.NewArrayWith(runtime.String("one"))}), runtime.ErrInvalidArgumentType},
			{"header none", "headers", runtime.NewObjectWith(map[string]runtime.Value{"X-Test": runtime.None}), runtime.ErrInvalidArgument},
			{"empty cookie", "cookies", runtime.NewObject(), runtime.ErrMissedArgument},
			{"cookie value missing", "cookies", runtime.NewObjectWith(map[string]runtime.Value{"name": runtime.String("session")}), runtime.ErrMissedArgument},
			{"cookie container", "cookies", runtime.String("secret"), runtime.ErrInvalidType},
			{"cookie list entry", "cookies", runtime.NewArrayWith(cookieMap("session", "probe"), runtime.True), runtime.ErrInvalidType},
			{"unknown option", "extra", runtime.True, runtime.ErrInvalidArgument},
			{"unknown viewport field", "viewport", runtime.NewObjectWith(map[string]runtime.Value{"extra": runtime.True}), runtime.ErrInvalidArgument},
		} {
			t.Run(function+"/"+tc.name, func(t *testing.T) {
				_, err := callPageOptions(t, function, runtime.NewObjectWith(map[string]runtime.Value{tc.key: tc.value}))
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("error = %v, want %v", err, tc.wantErr)
				}

				if strings.Contains(err.Error(), "secret") {
					t.Fatal("error exposes the input value")
				}
			})
		}
	}
}

func TestParseForwardsOrdinaryOptions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   runtime.Value
		options runtime.Value
		want    drivers.ParseParams
	}{
		{"defaults", runtime.String("original"), nil, drivers.ParseParams{Content: []byte("original")}},
		{"binary", runtime.Binary("original"), runtime.NewObject(), drivers.ParseParams{Content: []byte("original")}},
		{"options", runtime.String("original"), runtime.NewObjectWith(map[string]runtime.Value{
			"driver": runtime.String("recording"), "keepCookies": runtime.True,
			"viewport": runtime.NewObjectWith(map[string]runtime.Value{"width": runtime.Int(32), "height": runtime.Int(16)}),
			"content":  runtime.Binary("override"),
		}), drivers.ParseParams{Content: []byte("override"), KeepCookies: true, Viewport: &drivers.Viewport{Width: 32, Height: 16}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, drv := newRecordingDriver(t)
			args := []runtime.Value{tc.input}

			if tc.options != nil {
				args = append(args, tc.options)
			}

			_, err := Parse(ctx, args...)
			if err != nil {
				t.Fatal(err)
			}

			if drv.parseParams == nil || !reflect.DeepEqual(*drv.parseParams, tc.want) {
				t.Fatalf("Parse received %#v, want %+v", drv.parseParams, tc.want)
			}
		})
	}
}

func TestOpenForwardsOrdinaryOptions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options runtime.Value
		want    drivers.Params
	}{
		{"defaults", nil, drivers.Params{URL: "https://example.com"}},
		{"options", runtime.NewObjectWith(map[string]runtime.Value{
			"driver": runtime.String("recording"), "keepCookies": runtime.True,
			"userAgent": runtime.String("probe"), "charset": runtime.String("utf-8"),
		}), drivers.Params{URL: "https://example.com", KeepCookies: true, UserAgent: "probe", Charset: "utf-8"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, drv := newRecordingDriver(t)
			args := []runtime.Value{runtime.String("https://example.com")}

			if tc.options != nil {
				args = append(args, tc.options)
			}

			_, err := Open(ctx, args...)
			if err != nil {
				t.Fatal(err)
			}

			if drv.openParams == nil || !reflect.DeepEqual(*drv.openParams, tc.want) {
				t.Fatalf("Open received %#v, want %+v", drv.openParams, tc.want)
			}
		})
	}
}

func callPageOptions(t *testing.T, function string, options runtime.Value) (drivers.ParseParams, error) {
	t.Helper()

	ctx, drv := newRecordingDriver(t)
	var err error

	if function == "Open" {
		_, err = Open(ctx, runtime.String("https://example.com"), options)
	} else {
		_, err = Parse(ctx, runtime.String("<html></html>"), options)
	}

	if err != nil {
		if drv.openParams != nil || drv.parseParams != nil {
			t.Fatal("driver was called with invalid options")
		}

		return drivers.ParseParams{}, err
	}

	if function == "Open" && drv.openParams != nil {
		return drivers.ParseParams{Headers: drv.openParams.Headers, Cookies: drv.openParams.Cookies}, nil
	}

	if function == "Parse" && drv.parseParams != nil {
		return *drv.parseParams, nil
	}

	t.Fatal("driver was not called")

	return drivers.ParseParams{}, nil
}
