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
	// Polyfill for Go 1.10
	SameSite int

	// HTTPCookie HTTPCookie object
	HTTPCookie struct {
		Name     string    `json:"name"`
		Value    string    `json:"Value"`
		Path     string    `json:"path"`
		Domain   string    `json:"domain"`
		Expires  time.Time `json:"expires"`
		MaxAge   int       `json:"maxAge"`
		Secure   bool      `json:"secure"`
		HTTPOnly bool      `json:"HTTPOnly"`
		SameSite SameSite  `json:"sameSite"`
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

func (c HTTPCookie) Compare(other runtime.Value) int {
	oc, ok := other.(HTTPCookie)

	if !ok {
		return CompareTo(HTTPCookieType, other)
	}

	if c.Name != oc.Name {
		return strings.Compare(c.Name, oc.Name)
	}

	if c.Value != oc.Value {
		return strings.Compare(c.Value, oc.Value)
	}

	if c.Path != oc.Path {
		return strings.Compare(c.Path, oc.Path)
	}

	if c.Domain != oc.Domain {
		return strings.Compare(c.Domain, oc.Domain)
	}

	if c.Expires.After(oc.Expires) {
		return 1
	} else if c.Expires.Before(oc.Expires) {
		return -1
	}

	if c.MaxAge > oc.MaxAge {
		return 1
	} else if c.MaxAge < oc.MaxAge {
		return -1
	}

	if c.Secure && !oc.Secure {
		return 1
	} else if !c.Secure && oc.Secure {
		return -1
	}

	if c.HTTPOnly && !oc.HTTPOnly {
		return 1
	} else if !c.HTTPOnly && oc.HTTPOnly {
		return -1
	}

	if c.SameSite > oc.SameSite {
		return 1
	} else if c.SameSite < oc.SameSite {
		return -1
	}

	return 0
}

func (c HTTPCookie) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte(c.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(c.Name))
	h.Write([]byte(c.Value))
	h.Write([]byte(c.Path))
	h.Write([]byte(c.Domain))
	h.Write([]byte(c.Expires.String()))
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
		"name":      c.Name,
		"Value":     c.Value,
		"path":      c.Path,
		"domain":    c.Domain,
		"expires":   c.Expires,
		"max_age":   c.MaxAge,
		"secure":    c.Secure,
		"http_only": c.HTTPOnly,
		"same_site": c.SameSite.String(),
	}

	out, err := json.Marshal(v)

	if err != nil {
		return nil, err
	}

	return out, err
}

func (c HTTPCookie) Get(_ context.Context, key runtime.Value) (runtime.Value, error) {
	switch key.String() {
	case "name":
		return runtime.NewString(c.Name), nil
	case "Value":
		return runtime.NewString(c.Value), nil
	case "path":
		return runtime.NewString(c.Path), nil
	case "domain":
		return runtime.NewString(c.Domain), nil
	case "expires":
		return runtime.NewDateTime(c.Expires), nil
	case "maxAge":
		return runtime.NewInt(c.MaxAge), nil
	case "secure":
		return runtime.NewBoolean(c.Secure), nil
	case "httpOnly":
		return runtime.NewBoolean(c.HTTPOnly), nil
	case "sameSite":
		return runtime.NewString(c.SameSite.String()), nil
	default:
		return runtime.None, nil
	}
}
