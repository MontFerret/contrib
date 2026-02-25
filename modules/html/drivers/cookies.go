package drivers

import (
	"context"
	"strings"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type HTTPCookies struct {
	Data map[string]HTTPCookie
}

func NewHTTPCookies() *HTTPCookies {
	return &HTTPCookies{Data: make(map[string]HTTPCookie)}
}

func NewHTTPCookiesFrom(cookies *HTTPCookies) *HTTPCookies {
	return NewHTTPCookiesWith(cookies.Data)
}

func NewHTTPCookiesWith(values map[string]HTTPCookie) *HTTPCookies {
	data := make(map[string]HTTPCookie, len(values))

	for k, v := range values {
		data[k] = v
	}

	return &HTTPCookies{Data: values}
}

func (c *HTTPCookies) Hash() uint64 {
	//TODO implement me
	panic("implement me")
}

func (c *HTTPCookies) Type() runtime.Type {
	return HTTPCookiesType
}

func (c *HTTPCookies) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	cookie, exists := c.Data[key.String()]

	if !exists {
		return runtime.None, nil
	}

	return sdk.NewProxy(cookie), nil
}

func (c *HTTPCookies) Set(_ context.Context, key, value runtime.Value) error {
	switch v := value.(type) {
	case HTTPCookie:
		c.Data[key.String()] = v
	default:
		return runtime.TypeErrorOf(value, HTTPCookieType)
	}

	return nil
}

func (c *HTTPCookies) SetCookie(cookie HTTPCookie) {
	c.Data[cookie.Name] = cookie
}

func (c *HTTPCookies) Iterate(_ context.Context) (runtime.Iterator, error) {
	return sdk.NewMapIterator[string, HTTPCookie](c.Data), nil
}

func (c *HTTPCookies) String() string {
	var b strings.Builder

	b.WriteString("HTTPCookies{")

	for k, v := range c.Data {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v.String())
		b.WriteString(", ")
	}

	b.WriteString("}")

	return b.String()
}

func (c *HTTPCookies) Copy() runtime.Value {
	return NewHTTPCookiesWith(c.Data)
}
