package drivers

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const pkg = "html.drivers"

var (
	HTTPRequestType  = runtime.NewTypeFor[*HTTPRequest](pkg, "HTTPRequest")
	HTTPResponseType = runtime.NewTypeFor[*HTTPResponse](pkg, "HTTPResponse")
	HTTPHeadersType  = runtime.NewTypeFor[*HTTPHeaders](pkg, "HTTPHeaders")
	HTTPCookieType   = runtime.NewTypeFor[*HTTPCookie](pkg, "HTTPCookie")
	HTTPCookiesType  = runtime.NewTypeFor[*HTTPCookies](pkg, "HTTPCookies")
	HTMLElementType  = runtime.NewTypeFor[HTMLElement](pkg, "HTMLElement")
	HTMLDocumentType = runtime.NewTypeFor[HTMLDocument](pkg, "HTMLDocument")
	HTMLPageType     = runtime.NewTypeFor[HTMLPage](pkg, "HTMLPage")
)

// Comparison table of builtin types
var typeComparisonTable = map[runtime.Type]int{
	HTTPCookieType:   1,
	HTTPCookiesType:  2,
	HTTPRequestType:  3,
	HTTPResponseType: 4,
	HTMLElementType:  5,
	HTMLDocumentType: 6,
	HTMLPageType:     7,
}

func CompareTypes(first, second runtime.Value) int {
	typed1, ok1 := first.(runtime.Typed)
	typed2, ok2 := second.(runtime.Typed)

	if !ok1 || !ok2 {
		return -1
	}

	if typed1.Type() == typed2.Type() {
		return 0
	}

	return Compare(typed1.Type(), typed2.Type())
}

func Compare(first, second runtime.Type) int {
	f, ok := typeComparisonTable[first]

	if !ok {
		return -1
	}

	s, ok := typeComparisonTable[second]

	if !ok {
		return -1
	}

	if f == s {
		return 0
	}

	if f > s {
		return 1
	}

	return -1
}

func CompareTo(first runtime.Type, second runtime.Value) int {
	typed, ok := second.(runtime.Typed)

	if !ok {
		return -1
	}

	return Compare(first, typed.Type())
}
