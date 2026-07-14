package drivers

import (
	"context"
	"fmt"
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
	doc, ok := value.(HTMLDocument)
	if ok && !isNilValue(doc) {
		return nil, runtime.TypeErrorOf(value, HTMLElementType)
	}

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
	return toHTMLCapability(value, "content", func(value any) (ContentTarget, bool) {
		provider, ok := value.(contentTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsContentTarget(), true
	})
}

func ToAttributeTarget(value runtime.Value) (AttributeTarget, error) {
	return toHTMLCapability(value, "attribute", func(value any) (AttributeTarget, bool) {
		provider, ok := value.(attributeTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsAttributeTarget(), true
	})
}

func ToStyleTarget(value runtime.Value) (StyleTarget, error) {
	return toHTMLCapability(value, "style", func(value any) (StyleTarget, bool) {
		provider, ok := value.(styleTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsStyleTarget(), true
	})
}

func ToValueTarget(value runtime.Value) (ValueTarget, error) {
	return toHTMLCapability(value, "value", func(value any) (ValueTarget, bool) {
		provider, ok := value.(valueTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsValueTarget(), true
	})
}

func ToRelationTarget(value runtime.Value) (RelationTarget, error) {
	return toHTMLCapability(value, "relation", func(value any) (RelationTarget, bool) {
		provider, ok := value.(relationTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsRelationTarget(), true
	})
}

func ToInteractionTarget(value runtime.Value) (InteractionTarget, error) {
	return toHTMLCapability(value, "interaction", func(value any) (InteractionTarget, bool) {
		provider, ok := value.(interactionTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsInteractionTarget(), true
	})
}

func ToWaitTarget(value runtime.Value) (WaitTarget, error) {
	return toHTMLCapability(value, "wait", func(value any) (WaitTarget, bool) {
		provider, ok := value.(waitTargetProvider)
		if !ok {
			return nil, false
		}

		return provider.AsWaitTarget(), true
	})
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

func ToQuerySelector(ctx context.Context, value runtime.Value) (QuerySelector, error) {
	var qs QuerySelector

	switch v := value.(type) {
	case runtime.Map:
		if err := sdk.Decode(ctx, value, &qs); err != nil {
			return qs, fmt.Errorf("invalid selector map, expected keys 'Kind' and 'Value': %w", err)
		}

		return qs, nil
	case runtime.String:
		return NewCSSSelector(v), nil
	case *sdk.HostValue[QuerySelector]:
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

func toHTMLCapability[T any](value runtime.Value, capability string, provider func(any) (T, bool)) (T, error) {
	var zero T

	switch value.(type) {
	case HTMLPage, HTMLDocument, HTMLElement:
		return asCapability(value, capability, provider)
	default:
		return zero, runtime.TypeErrorOf(value, HTMLPageType, HTMLDocumentType, HTMLElementType)
	}
}

func toDocumentCapability[T any](value runtime.Value, capability string) (T, error) {
	var zero T

	switch v := value.(type) {
	case HTMLPage:
		return asCapability[T](v.GetMainFrame(), capability, nil)
	case HTMLDocument:
		return asCapability[T](v, capability, nil)
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

	return asCapability[T](page, capability, nil)
}

func asCapability[T any](value any, capability string, provider func(any) (T, bool)) (T, error) {
	var zero T

	target, ok := value.(T)
	if ok && !isNilValue(target) {
		return target, nil
	}

	if provider != nil {
		target, ok = provider(value)
		if ok && !isNilValue(target) {
			return target, nil
		}
	}

	return zero, runtime.Errorf(runtime.ErrNotSupported, "%s capability", capability)
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
