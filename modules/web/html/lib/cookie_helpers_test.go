package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func TestParseCookiesValueFromSingleCookie(t *testing.T) {
	t.Parallel()

	cookies, err := parseCookiesValue(context.Background(), cookieMap("session", "abc123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedCookie(t, cookies, "session", "abc123")
}

func TestParseCookiesValueFromList(t *testing.T) {
	t.Parallel()

	input := runtime.NewArrayWith(
		cookieMap("session", "abc123"),
		drivers.HTTPCookie{Name: "theme", Value: "dark"},
	)

	cookies, err := parseCookiesValue(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedCookie(t, cookies, "session", "abc123")
	assertParsedCookie(t, cookies, "theme", "dark")
}

func TestParseCookiesValueFromCookieCollection(t *testing.T) {
	t.Parallel()

	input := drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{
		"session": {Name: "session", Value: "abc123"},
	})

	cookies, err := parseCookiesValue(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedCookie(t, cookies, "session", "abc123")
}

func TestParseCookiesValueFromProxiedCookieCollection(t *testing.T) {
	t.Parallel()

	input := drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{
		"session": {Name: "session", Value: "abc123"},
	})

	cookies, err := parseCookiesValue(context.Background(), sdk.NewProxy(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedCookie(t, cookies, "session", "abc123")
}

func TestParseCookiesValueAcceptsLegacyCookieMapFields(t *testing.T) {
	t.Parallel()

	input := runtime.NewObjectWith(map[string]runtime.Value{
		"Name":      runtime.NewString("session"),
		"Value":     runtime.NewString("abc123"),
		"Path":      runtime.NewString("/"),
		"max_age":   runtime.NewInt(60),
		"same_site": runtime.NewString("strict"),
		"http_only": runtime.True,
		"Secure":    runtime.True,
	})

	cookies, err := parseCookiesValue(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertParsedCookie(t, cookies, "session", "abc123")

	cookie := cookies.Data["session"]
	if cookie.Path != "/" {
		t.Fatalf("expected legacy path alias to be parsed, got %q", cookie.Path)
	}

	if cookie.MaxAge != 60 {
		t.Fatalf("expected legacy max_age alias to be parsed, got %d", cookie.MaxAge)
	}

	if cookie.SameSite != drivers.SameSiteStrictMode {
		t.Fatalf("expected legacy same_site alias to be parsed, got %v", cookie.SameSite)
	}

	if !cookie.HTTPOnly {
		t.Fatal("expected legacy http_only alias to be parsed")
	}

	if !cookie.Secure {
		t.Fatal("expected legacy secure alias to be parsed")
	}
}

func TestParseCookiesValueRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	if _, err := parseCookiesValue(context.Background(), runtime.None); !errors.Is(err, runtime.ErrMissedArgument) {
		t.Fatalf("expected missing cookies error, got %v", err)
	}

	if _, err := parseCookiesValue(context.Background(), runtime.NewString("invalid")); err == nil {
		t.Fatal("expected invalid cookie type error")
	}

	if _, err := parseCookiesValue(context.Background(), runtime.NewObject()); !errors.Is(err, runtime.ErrMissedArgument) {
		t.Fatalf("expected missing cookie name error, got %v", err)
	}
}

func cookieMap(name, value string) runtime.Map {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"name":  runtime.NewString(name),
		"value": runtime.NewString(value),
	})
}

func assertParsedCookie(t *testing.T, cookies *drivers.HTTPCookies, name, value string) {
	t.Helper()

	if cookies == nil {
		t.Fatal("expected cookies, got nil")
	}

	cookie, ok := cookies.Data[name]
	if !ok {
		t.Fatalf("expected cookie %q to be parsed, got %#v", name, cookies.Data)
	}

	if cookie.Value != value {
		t.Fatalf("expected cookie %q value %q, got %q", name, value, cookie.Value)
	}
}
