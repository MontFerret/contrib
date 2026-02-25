package html

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/drivers/cdp"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type PageLoadParams struct {
	drivers.Params
	Driver  string        `json:"driver"`
	Timeout time.Duration `json:"timeout"`
}

// DOCUMENT opens an HTML page by a given url.
// By default, loads a page by http call - resulted page does not support any interactions.
// @param {Object} [params] - An object containing the following properties :
// @param {String} [params.driver] - Driver name to use.
// @param {Int} [params.timeout=60000] - Page load timeout.
// @param {String} [params.userAgent] - Custom user agent.
// @param {Boolean} [params.keepCookies=False] - Boolean value indicating whether to use cookies from previous sessions i.e. not to open a page in the Incognito mode.
// @param {Object[] | Object} [params.cookies] - Set of HTTP cookies to use during page loading.
// @param {String} params.cookies.*.name - Cookie name.
// @param {String} params.cookies.*.value - Cookie value.
// @param {String} params.cookies.*.path - Cookie path.
// @param {String} params.cookies.*.domain - Cookie domain.
// @param {Int} [params.cookies.*.maxAge] - Cookie max age.
// @param {String|DateTime} [params.cookies.*.expires] - Cookie expiration date time.
// @param {String} [params.cookies.*.sameSite] - Cookie cross-origin policy.
// @param {Boolean} [params.cookies.*.httpOnly=false] - Cookie cannot be accessed through client side script.
// @param {Boolean} [params.cookies.*.secure=false] - Cookie sent to the server only with an encrypted request over the HTTPS protocol.
// @param {Object} [params.headers] - Set of HTTP headers to use during page loading.
// @param {Object} [params.ignore] - Set of parameters to ignore some page functionality or behavior.
// @param {Object[]} [params.ignore.resources] - Collection of rules to ignore resources during page load and navigation.
// @param {String} [params.ignore.resources.*.url] - Resource url pattern. If set, requests for matching urls will be blocked. Wildcards ('*' -> zero or more, '?' -> exactly one) are allowed. Escape character is backslash. Omitting is equivalent to "*".
// @param {String} [params.ignore.resources.*.type] - Resource type. If set, requests for matching resource types will be blocked.
// @param {Object[]} [params.ignore.statusCodes] - Collection of rules to ignore certain HTTP codes that can cause failures.
// @param {String} [params.ignore.statusCodes.*.url] - Url pattern. If set, codes for matching urls will be ignored. Wildcards ('*' -> zero or more, '?' -> exactly one) are allowed. Escape character is backslash. Omitting is equivalent to "*".
// @param {Int} [params.ignore.statusCodes.*.code] - HTTP code to ignore.
// @param {Object} [params.viewport] - Viewport params.
// @param {Int} [params.viewport.height] - Viewport height.
// @param {Int} [params.viewport.width] - Viewport width.
// @param {Float} [params.viewport.scaleFactor] - Viewport scale factor.
// @param {Boolean} [params.viewport.mobile] - Value that indicates whether to emulate mobile device.
// @param {Boolean} [params.viewport.landscape] - Value that indicates whether to render a page in landscape position.
// @param {String} [params.charset] - (only HTTPDriver) Source charset content to convert UTF-8.
// @return {HTMLPage} - Loaded HTML page.
func Open(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	url, err := runtime.CastString(args[0])

	if err != nil {
		return runtime.None, err
	}

	var params PageLoadParams

	if len(args) == 1 {
		params = newDefaultDocLoadParams(url)
	} else {
		p, err := newPageLoadParams(url, args[1])

		if err != nil {
			return runtime.None, err
		}

		params = p
	}

	ctx, cancel := context.WithTimeout(ctx, params.Timeout)
	defer cancel()

	drv, err := drivers.FromContext(ctx, params.Driver)

	if err != nil {
		return runtime.None, err
	}

	return drv.Open(ctx, params.Params)
}

func newDefaultDocLoadParams(url runtime.String) PageLoadParams {
	return PageLoadParams{
		Params: drivers.Params{
			URL: url.String(),
		},
		Timeout: drivers.DefaultPageLoadTimeout * time.Millisecond,
	}
}

func newPageLoadParams(url runtime.String, arg runtime.Value) (PageLoadParams, error) {
	if err := runtime.ValidateType(arg, runtime.TypeBoolean, runtime.TypeString, runtime.TypeMap); err != nil {
		return PageLoadParams{}, err
	}

	res := newDefaultDocLoadParams(url)

	switch argt := arg.(type) {
	case runtime.Map:
		if err := sdk.Decode(argt, &res); err != nil {
			return PageLoadParams{}, err
		}
	case runtime.String:
		res.Driver = arg.(runtime.String).String()
	case runtime.Boolean:
		b := arg.(runtime.Boolean)

		// fallback
		if b {
			res.Driver = cdp.DriverName
		}
	}

	return res, nil
}

//func parseCookieObject(obj *runtime.Object) (*drivers.HTTPCookies, error) {
//	if obj == nil {
//		return nil, errors.Wrap(runtime.ErrMissedArgument, "cookies")
//	}
//
//	var err error
//	res := drivers.NewHTTPCookies()
//
//	obj.ForEach(func(value runtime.Value, _ string) bool {
//		cookie, e := parseCookie(value)
//
//		if e != nil {
//			err = e
//
//			return false
//		}
//
//		res.Set(cookie)
//
//		return true
//	})
//
//	return res, err
//}
//
//func parseCookieArray(arr *runtime.Array) (*drivers.HTTPCookies, error) {
//	if arr == nil {
//		return nil, errors.Wrap(runtime.ErrMissedArgument, "cookies")
//	}
//
//	var err error
//	res := drivers.NewHTTPCookies()
//
//	arr.ForEach(func(value runtime.Value, _ int) bool {
//		cookie, e := parseCookie(value)
//
//		if e != nil {
//			err = e
//
//			return false
//		}
//
//		res.Set(cookie)
//
//		return true
//	})
//
//	return res, err
//}
//
//func parseCookie(value runtime.Value) (drivers.HTTPCookie, error) {
//	err := runtime.ValidateType(value, runtime.TypeObject, drivers.HTTPCookieType)
//
//	if err != nil {
//		return drivers.HTTPCookie{}, err
//	}
//
//	if value.Type() == drivers.HTTPCookieType {
//		return value.(drivers.HTTPCookie), nil
//	}
//
//	co := value.(*runtime.Object)
//
//	cookie := drivers.HTTPCookie{
//		Name:   co.MustGet("name").String(),
//		Value:  co.MustGet("value").String(),
//		Path:   co.MustGet("path").String(),
//		Domain: co.MustGet("domain").String(),
//	}
//
//	maxAge, exists := co.Get("maxAge")
//
//	if exists {
//		if err = runtime.ValidateType(maxAge, runtime.TypeInt); err != nil {
//			return drivers.HTTPCookie{}, err
//		}
//
//		cookie.MaxAge = int(maxAge.(runtime.Int))
//	}
//
//	expires, exists := co.Get("expires")
//
//	if exists {
//		if err = runtime.ValidateType(expires, runtime.TypeDateTime, runtime.TypeString); err != nil {
//			return drivers.HTTPCookie{}, err
//		}
//
//		if expires.Type() == runtime.TypeDateTime {
//			cookie.Expires = expires.(runtime.DateTime).Unwrap().(time.Time)
//		} else {
//			t, err := time.Parse(runtime.DefaultTimeLayout, expires.String())
//
//			if err != nil {
//				return drivers.HTTPCookie{}, err
//			}
//
//			cookie.Expires = t
//		}
//	}
//
//	sameSite, exists := co.Get("sameSite")
//
//	if exists {
//		sameSite := strings.ToLower(sameSite.String())
//
//		switch sameSite {
//		case "lax":
//			cookie.SameSite = drivers.SameSiteLaxMode
//		case "strict":
//			cookie.SameSite = drivers.SameSiteStrictMode
//		default:
//			cookie.SameSite = drivers.SameSiteDefaultMode
//		}
//	}
//
//	httpOnly, exists := co.Get("httpOnly")
//
//	if exists {
//		if err = runtime.ValidateType(httpOnly, runtime.TypeBoolean); err != nil {
//			return drivers.HTTPCookie{}, err
//		}
//
//		cookie.HTTPOnly = bool(httpOnly.(runtime.Boolean))
//	}
//
//	secure, exists := co.Get("secure")
//
//	if exists {
//		if err = runtime.ValidateType(secure, runtime.TypeBoolean); err != nil {
//			return drivers.HTTPCookie{}, err
//		}
//
//		cookie.Secure = bool(secure.(runtime.Boolean))
//	}
//
//	return cookie, err
//}
//
//func parseHeader(headers *runtime.Object) *drivers.HTTPHeaders {
//	res := drivers.NewHTTPHeaders()
//
//	headers.ForEach(func(value runtime.Value, key string) bool {
//		if value.Type() == runtime.TypeArray {
//			value := value.(*runtime.Array)
//
//			keyValues := make([]string, 0, value.Length())
//
//			value.ForEach(func(v runtime.Value, _ int) bool {
//				keyValues = append(keyValues, v.String())
//
//				return true
//			})
//
//			res.SetArr(key, keyValues)
//		} else {
//			res.Set(key, value.String())
//		}
//
//		return true
//	})
//
//	return res
//}
//
//func parseViewport(value runtime.Value) (*drivers.Viewport, error) {
//	if err := runtime.ValidateType(value, runtime.TypeObject); err != nil {
//		return nil, err
//	}
//
//	res := &drivers.Viewport{}
//
//	viewport := value.(*runtime.Object)
//
//	width, exists := viewport.Get(runtime.NewString("width"))
//
//	if exists {
//		if err := runtime.ValidateType(width, runtime.TypeInt); err != nil {
//			return nil, err
//		}
//
//		res.Width = int(runtime.ToInt(width))
//	}
//
//	height, exists := viewport.Get(runtime.NewString("height"))
//
//	if exists {
//		if err := runtime.ValidateType(height, runtime.TypeInt); err != nil {
//			return nil, err
//		}
//
//		res.Height = int(runtime.ToInt(height))
//	}
//
//	mobile, exists := viewport.Get(runtime.NewString("mobile"))
//
//	if exists {
//		res.Mobile = bool(runtime.ToBoolean(mobile))
//	}
//
//	landscape, exists := viewport.Get(runtime.NewString("landscape"))
//
//	if exists {
//		res.Landscape = bool(runtime.ToBoolean(landscape))
//	}
//
//	scaleFactor, exists := viewport.Get(runtime.NewString("scaleFactor"))
//
//	if exists {
//		res.ScaleFactor = float64(runtime.ToFloat(scaleFactor))
//	}
//
//	return res, nil
//}
//
//func parseIgnore(value runtime.Value) (*drivers.Ignore, error) {
//	if err := runtime.ValidateType(value, runtime.TypeObject); err != nil {
//		return nil, err
//	}
//
//	res := &drivers.Ignore{}
//
//	ignore := value.(*runtime.Object)
//
//	resources, exists := ignore.Get("resources")
//
//	if exists {
//		if err := runtime.ValidateType(resources, runtime.TypeArray); err != nil {
//			return nil, err
//		}
//
//		resources := resources.(*runtime.Array)
//
//		res.Resources = make([]drivers.ResourceFilter, 0, resources.Length())
//
//		var e error
//
//		resources.ForEach(func(el runtime.Value, _ int) bool {
//			if e = runtime.ValidateType(el, runtime.TypeObject); e != nil {
//				return false
//			}
//
//			pattern := el.(*runtime.Object)
//
//			url, urlExists := pattern.Get("url")
//			resType, resTypeExists := pattern.Get("type")
//
//			// ignore element
//			if !urlExists && !resTypeExists {
//				return true
//			}
//
//			res.Resources = append(res.Resources, drivers.ResourceFilter{
//				URL:  url.String(),
//				Type: resType.String(),
//			})
//
//			return true
//		})
//
//		if e != nil {
//			return nil, e
//		}
//	}
//
//	statusCodes, exists := ignore.Get("statusCodes")
//
//	if exists {
//		if err := runtime.ValidateType(statusCodes, runtime.TypeArray); err != nil {
//			return nil, err
//		}
//
//		statusCodes := statusCodes.(*runtime.Array)
//
//		res.StatusCodes = make([]drivers.StatusCodeFilter, 0, statusCodes.Length())
//
//		var e error
//
//		statusCodes.ForEach(func(el runtime.Value, _ int) bool {
//			if e = runtime.ValidateType(el, runtime.TypeObject); e != nil {
//				return false
//			}
//
//			pattern := el.(*runtime.Object)
//
//			url := pattern.MustGetOr("url", runtime.NewString(""))
//			code, codeExists := pattern.Get("code")
//
//			// ignore element
//			if !codeExists {
//				e = errors.New("http code is required")
//				return false
//			}
//
//			res.StatusCodes = append(res.StatusCodes, drivers.StatusCodeFilter{
//				URL:  url.String(),
//				Code: int(runtime.ToInt(code)),
//			})
//
//			return true
//		})
//	}
//
//	return res, nil
//}
