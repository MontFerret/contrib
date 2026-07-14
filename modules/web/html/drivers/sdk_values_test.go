package drivers_test

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func TestQuerySelectorHostValueUnwrapsWithoutExtraCapabilities(t *testing.T) {
	t.Parallel()

	expected := drivers.NewXPathSelector("//a")
	value := sdk.NewHostValueWithType(drivers.TypeQuerySelector, expected)

	actual, err := drivers.ToQuerySelector(context.Background(), value)
	if err != nil {
		t.Fatalf("unexpected selector conversion error: %v", err)
	}
	if actual != expected {
		t.Fatalf("unexpected selector: got %#v, want %#v", actual, expected)
	}

	if _, ok := any(value).(runtime.Iterable); ok {
		t.Fatal("opaque selector must not be iterable")
	}
	if _, ok := any(value).(runtime.KeyReadable); ok {
		t.Fatal("opaque selector must not be key-readable")
	}
	if _, ok := any(value).(runtime.Queryable); ok {
		t.Fatal("opaque selector must not be queryable")
	}
}

func TestHTTPValuesRetainNativeCapabilities(t *testing.T) {
	t.Parallel()

	headers := drivers.NewHTTPHeadersWith(map[string][]string{"Content-Type": {"text/plain"}})
	request := &drivers.HTTPRequest{Headers: headers}
	headerValue, err := request.Get(context.Background(), runtime.NewString("headers"))
	if err != nil {
		t.Fatalf("unexpected header lookup error: %v", err)
	}
	if headerValue != headers {
		t.Fatal("expected request to return the native headers value")
	}
	if _, ok := headerValue.(runtime.KeyReadable); !ok {
		t.Fatal("expected headers to remain key-readable")
	}
	if _, ok := headerValue.(runtime.KeyWritable); !ok {
		t.Fatal("expected headers to remain key-writable")
	}
	if _, ok := headerValue.(runtime.Iterable); !ok {
		t.Fatal("expected headers to remain iterable")
	}

	cookie := drivers.HTTPCookie{Name: "session", Value: "token"}
	cookies := drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{"session": cookie})
	cookieValue, err := cookies.Get(context.Background(), runtime.NewString("session"))
	if err != nil {
		t.Fatalf("unexpected cookie lookup error: %v", err)
	}
	if _, ok := cookieValue.(drivers.HTTPCookie); !ok {
		t.Fatalf("expected native cookie value, got %T", cookieValue)
	}
	if _, ok := cookieValue.(runtime.KeyReadable); !ok {
		t.Fatal("expected cookie to remain key-readable")
	}
	if _, ok := cookieValue.(runtime.Iterable); ok {
		t.Fatal("cookie must not gain iterable capability")
	}
}
