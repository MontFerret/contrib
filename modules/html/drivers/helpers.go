package drivers

import (
	"errors"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func ToPage(value runtime.Value) (HTMLPage, error) {
	err := runtime.ValidateType(value, HTMLPageType)

	if err != nil {
		return nil, err
	}

	return value.(HTMLPage), nil
}

func ToDocument(value runtime.Value) (HTMLDocument, error) {
	switch v := value.(type) {
	case HTMLPage:
		return v.GetMainFrame(), nil
	case HTMLDocument:
		return v, nil
	default:
		return nil, runtime.TypeErrorOf(
			value,
			HTMLPageType,
			HTMLDocumentType,
		)
	}
}

func ToElement(value runtime.Value) (HTMLElement, error) {
	switch v := value.(type) {
	case HTMLPage:
		return v.GetMainFrame().GetElement(), nil
	case HTMLDocument:
		return v.GetElement(), nil
	case HTMLElement:
		return v, nil
	default:
		return nil, runtime.TypeErrorOf(
			value,
			HTMLPageType,
			HTMLDocumentType,
			HTMLElementType,
		)
	}
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
		for key, _ := range opts.Headers.Data {
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
		for name, _ := range opts.Cookies.Data {
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
