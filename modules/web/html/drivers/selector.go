package drivers

import (
	"reflect"

	"github.com/goccy/go-json"

	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	QuerySelectorKind int

	QuerySelector struct {
		Value runtime.String    `json:"Value"`
		Kind  QuerySelectorKind `json:"Kind"`
	}
)

var (
	TypeQuerySelector = runtime.NewType(reflect.TypeOf(QuerySelector{}).PkgPath(), reflect.TypeOf(QuerySelector{}).Name(), func(value runtime.Value) bool {
		switch value.(type) {
		case *runtime.Box[QuerySelector], *sdk.Proxy[QuerySelector]:
			return true
		default:
			return false
		}
	})
)

const (
	UnknownSelector QuerySelectorKind = iota
	CSSSelector
	XPathSelector
)

var (
	qsvStr = map[QuerySelectorKind]string{
		UnknownSelector: "unknown",
		CSSSelector:     "css",
		XPathSelector:   "xpath",
	}
)

func (v QuerySelectorKind) String() string {
	str, found := qsvStr[v]

	if found {
		return str
	}

	return qsvStr[UnknownSelector]
}

func NewCSSSelector(value runtime.String) QuerySelector {
	return QuerySelector{
		Kind:  CSSSelector,
		Value: value,
	}
}

func NewXPathSelector(value runtime.String) QuerySelector {
	return QuerySelector{
		Kind:  XPathSelector,
		Value: value,
	}
}

func (q QuerySelector) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"kind":  q.Kind.String(),
		"value": q.Value.String(),
	})
}

func (q QuerySelector) String() string {
	return q.Value.String()
}
