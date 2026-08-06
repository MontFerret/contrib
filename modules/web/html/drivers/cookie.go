package drivers

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	// SameSite Polyfill for Go 1.10
	SameSite int

	// HTTPCookie HTTPCookie object
	HTTPCookie struct {
		Expires  time.Time `json:"expires"`
		Name     string    `json:"name"`
		Value    string    `json:"value"`
		Path     string    `json:"path"`
		Domain   string    `json:"domain"`
		MaxAge   int       `json:"maxAge"`
		SameSite SameSite  `json:"sameSite"`
		Secure   bool      `json:"secure"`
		HTTPOnly bool      `json:"httpOnly"`
	}
)

const (
	SameSiteDefaultMode SameSite = iota + 1
	SameSiteLaxMode
	SameSiteStrictMode
)

func (s SameSite) String() string {
	switch s {
	case SameSiteLaxMode:
		return "Lax"
	case SameSiteStrictMode:
		return "Strict"
	default:
		return ""
	}
}

func (c HTTPCookie) Type() runtime.Type {
	return HTTPCookieType
}

func (c HTTPCookie) String() string {
	return fmt.Sprintf("%s=%s", c.Name, c.Value)
}

func (c HTTPCookie) Equal(ctx context.Context, other runtime.Value) (bool, error) {
	if _, err := checkComparisonContext(ctx); err != nil {
		return false, err
	}

	oc, ok := cookieValue(other)
	if !ok {
		return false, nil
	}

	comparison, err := c.compare(ctx, oc)

	return comparison == runtime.Equal, err
}

func (c HTTPCookie) Compare(ctx context.Context, other runtime.Value) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	oc, ok := cookieValue(other)
	if !ok {
		return invalidComparison(c, other)
	}

	return c.compare(ctx, oc)
}

func (c HTTPCookie) compare(ctx context.Context, oc HTTPCookie) (runtime.Ordering, error) {
	if comparison, err := checkComparisonContext(ctx); err != nil {
		return comparison, err
	}

	if c.Name != oc.Name {
		return runtime.Ordering(strings.Compare(c.Name, oc.Name)), nil
	}

	if c.Value != oc.Value {
		return runtime.Ordering(strings.Compare(c.Value, oc.Value)), nil
	}

	if c.Path != oc.Path {
		return runtime.Ordering(strings.Compare(c.Path, oc.Path)), nil
	}

	if c.Domain != oc.Domain {
		return runtime.Ordering(strings.Compare(c.Domain, oc.Domain)), nil
	}

	if c.Expires.After(oc.Expires) {
		return runtime.Greater, nil
	} else if c.Expires.Before(oc.Expires) {
		return runtime.Less, nil
	}

	if c.MaxAge > oc.MaxAge {
		return runtime.Greater, nil
	} else if c.MaxAge < oc.MaxAge {
		return runtime.Less, nil
	}

	if c.Secure && !oc.Secure {
		return runtime.Greater, nil
	} else if !c.Secure && oc.Secure {
		return runtime.Less, nil
	}

	if c.HTTPOnly && !oc.HTTPOnly {
		return runtime.Greater, nil
	} else if !c.HTTPOnly && oc.HTTPOnly {
		return runtime.Less, nil
	}

	if c.SameSite > oc.SameSite {
		return runtime.Greater, nil
	} else if c.SameSite < oc.SameSite {
		return runtime.Less, nil
	}

	return runtime.Equal, nil
}

func cookieValue(value runtime.Value) (HTTPCookie, bool) {
	switch cookie := value.(type) {
	case HTTPCookie:
		return cookie, true
	case *HTTPCookie:
		if cookie != nil {
			return *cookie, true
		}
	}

	return HTTPCookie{}, false
}

func (c HTTPCookie) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte(c.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(c.Name))
	h.Write([]byte(c.Value))
	h.Write([]byte(c.Path))
	h.Write([]byte(c.Domain))
	h.Write([]byte(c.Expires.UTC().Format(time.RFC3339Nano)))
	h.Write([]byte(strconv.Itoa(c.MaxAge)))
	h.Write([]byte(fmt.Sprintf("%t", c.Secure)))
	h.Write([]byte(fmt.Sprintf("%t", c.HTTPOnly)))
	h.Write([]byte(c.SameSite.String()))

	return h.Sum64()
}

func (c HTTPCookie) Copy() runtime.Value {
	cop := c
	return &cop
}

func (c HTTPCookie) MarshalJSON() ([]byte, error) {
	v := map[string]any{
		"name":     c.Name,
		"value":    c.Value,
		"path":     c.Path,
		"domain":   c.Domain,
		"expires":  c.Expires,
		"maxAge":   c.MaxAge,
		"secure":   c.Secure,
		"httpOnly": c.HTTPOnly,
		"sameSite": c.SameSite.String(),
	}

	out, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return out, err
}

func (c HTTPCookie) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	switch key.String() {
	case "name", "Name":
		return runtime.NewString(c.Name), nil
	case "value", "Value":
		return runtime.NewString(c.Value), nil
	case "path", "Path":
		return runtime.NewString(c.Path), nil
	case "domain", "Domain":
		return runtime.NewString(c.Domain), nil
	case "expires", "Expires":
		return runtime.NewDateTime(c.Expires), nil
	case "maxAge", "MaxAge":
		return runtime.NewInt(c.MaxAge), nil
	case "secure", "Secure":
		return runtime.NewBoolean(c.Secure), nil
	case "httpOnly", "HTTPOnly":
		return runtime.NewBoolean(c.HTTPOnly), nil
	case "sameSite", "SameSite":
		return runtime.NewString(c.SameSite.String()), nil
	default:
		return runtime.None, nil
	}
}
