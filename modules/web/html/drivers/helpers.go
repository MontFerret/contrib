package drivers

import (
	"errors"
	"reflect"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func ToPage(value runtime.Value) (HTMLPage, error) {
	page, ok := value.(HTMLPage)
	if !ok || isNilValue(page) {
		return nil, runtime.TypeErrorOf(value, HTMLPageType)
	}

	return page, nil
}

func ToDocument(value runtime.Value) (HTMLDocument, error) {
	doc, ok := value.(HTMLDocument)
	if !ok || isNilValue(doc) {
		return nil, runtime.TypeErrorOf(value, HTMLDocumentType)
	}

	return doc, nil
}

func ToElement(value runtime.Value) (HTMLElement, error) {
	el, ok := value.(HTMLElement)
	if !ok || isNilValue(el) {
		return nil, runtime.TypeErrorOf(value, HTMLElementType)
	}

	return el, nil
}

func ToQueryTarget(value runtime.Value) (QueryTarget, error) {
	switch v := value.(type) {
	case HTMLPage:
		return v.GetMainFrame(), nil
	case HTMLDocument:
		return v, nil
	case HTMLElement:
		return v, nil
	default:
		return nil, runtime.TypeErrorOf(value, HTMLPageType, HTMLDocumentType, HTMLElementType)
	}
}

func ToContentTarget(value runtime.Value) (ContentTarget, error) {
	return toHTMLCapability[ContentTarget](value, "content")
}

func ToAttributeTarget(value runtime.Value) (AttributeTarget, error) {
	return toHTMLCapability[AttributeTarget](value, "attribute")
}

func ToStyleTarget(value runtime.Value) (StyleTarget, error) {
	return toHTMLCapability[StyleTarget](value, "style")
}

func ToValueTarget(value runtime.Value) (ValueTarget, error) {
	return toHTMLCapability[ValueTarget](value, "value")
}

func ToRelationTarget(value runtime.Value) (RelationTarget, error) {
	return toHTMLCapability[RelationTarget](value, "relation")
}

func ToInteractionTarget(value runtime.Value) (InteractionTarget, error) {
	return toHTMLCapability[InteractionTarget](value, "interaction")
}

func ToWaitTarget(value runtime.Value) (WaitTarget, error) {
	return toHTMLCapability[WaitTarget](value, "wait")
}

func ToDocumentViewportTarget(value runtime.Value) (DocumentViewportTarget, error) {
	return toDocumentCapability[DocumentViewportTarget](value, "document viewport")
}

func ToPageCookieReader(value runtime.Value) (PageCookieReader, error) {
	return toPageCapability[PageCookieReader](value, "page cookies")
}

func ToPageCookieTarget(value runtime.Value) (PageCookieTarget, error) {
	return toPageCapability[PageCookieTarget](value, "page cookies")
}

func ToPageResponseTarget(value runtime.Value) (PageResponseTarget, error) {
	return toPageCapability[PageResponseTarget](value, "page response")
}

func ToPageSnapshotTarget(value runtime.Value) (PageSnapshotTarget, error) {
	return toPageCapability[PageSnapshotTarget](value, "page snapshot")
}

func ToPageNavigationTarget(value runtime.Value) (PageNavigationTarget, error) {
	return toPageCapability[PageNavigationTarget](value, "page navigation")
}

func ToQuerySelector(value runtime.Value) (QuerySelector, error) {
	var qs QuerySelector

	switch v := value.(type) {
	case runtime.Map:
		if err := sdk.Decode(value, &qs); err != nil {
			return qs, errors.New("invalid selector map, expected keys 'Kind' and 'Value'")
		}

		return qs, nil
	case runtime.String:
		return NewCSSSelector(v), nil
	case *sdk.Proxy[QuerySelector]:
		return v.Target(), nil
	case *runtime.Box[QuerySelector]:
		return v.Value, nil
	default:
		return qs, runtime.TypeErrorOf(value, runtime.TypeMap, runtime.TypeString)
	}
}

func SetDefaultParams(opts *Options, params Params) Params {
	if params.Headers == nil && opts.Headers != nil {
		params.Headers = NewHTTPHeaders()
	}

	// set default headers
	if opts.Headers != nil {
		for key := range opts.Headers.Data {
			val := params.Headers.Data.Get(key)

			// do not override user's set Data
			if val == "" {
				params.Headers.Data.Set(key, val)
			}
		}
	}

	if params.Cookies == nil && opts.Cookies != nil {
		params.Cookies = NewHTTPCookies()
	}

	// set default cookies
	if opts.Cookies != nil {
		for name := range opts.Cookies.Data {
			_, exists := params.Cookies.Data[name]

			// do not override user's set Data
			if !exists {
				params.Cookies.Data[name] = opts.Cookies.Data[name]
			}
		}
	}

	// set default user agent
	if opts.UserAgent != "" && params.UserAgent == "" {
		params.UserAgent = opts.UserAgent
	}

	return params
}

func toHTMLCapability[T any](value runtime.Value, capability string) (T, error) {
	var zero T

	switch value.(type) {
	case HTMLPage, HTMLDocument, HTMLElement:
		return asCapability[T](value, capability)
	default:
		return zero, runtime.TypeErrorOf(value, HTMLPageType, HTMLDocumentType, HTMLElementType)
	}
}

func toDocumentCapability[T any](value runtime.Value, capability string) (T, error) {
	var zero T

	switch v := value.(type) {
	case HTMLPage:
		return asCapability[T](v.GetMainFrame(), capability)
	case HTMLDocument:
		return asCapability[T](v, capability)
	default:
		return zero, runtime.TypeErrorOf(value, HTMLPageType, HTMLDocumentType)
	}
}

func toPageCapability[T any](value runtime.Value, capability string) (T, error) {
	var zero T

	page, err := ToPage(value)
	if err != nil {
		return zero, err
	}

	return asCapability[T](page, capability)
}

func asCapability[T any](value any, capability string) (T, error) {
	var zero T

	target, ok := value.(T)
	if !ok || isNilValue(target) {
		return zero, runtime.Errorf(runtime.ErrNotSupported, "%s capability", capability)
	}

	return target, nil
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}

	ref := reflect.ValueOf(value)
	switch ref.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return ref.IsNil()
	default:
		return false
	}
}
