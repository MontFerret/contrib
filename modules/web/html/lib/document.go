package lib

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

type (
	PageLoadParams struct {
		Driver string `json:"driver"`
		drivers.Params
		Timeout time.Duration `json:"timeout"`
	}

	pageLoadParamsInput struct {
		Driver      *string              `json:"driver"`
		Timeout     *time.Duration       `json:"timeout"`
		UserAgent   *string              `json:"userAgent"`
		KeepCookies *bool                `json:"keepCookies"`
		Cookies     runtime.Value        `json:"cookies"`
		Headers     *drivers.HTTPHeaders `json:"headers"`
		Viewport    *drivers.Viewport    `json:"viewport"`
		Ignore      *drivers.Ignore      `json:"ignore"`
		Charset     *string              `json:"charset"`
	}
)

// Open opens an HTML page by a given URL.
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
		p, err := newPageLoadParams(ctx, url, args[1])

		if err != nil {
			return runtime.None, err
		}

		params = p
	}

	ctx, cancel := context.WithTimeout(ctx, params.Timeout*time.Millisecond)
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

func newPageLoadParams(ctx context.Context, url runtime.String, arg runtime.Value) (PageLoadParams, error) {
	if err := runtime.ValidateType(arg, runtime.TypeBoolean, runtime.TypeString, runtime.TypeMap); err != nil {
		return PageLoadParams{}, err
	}

	res := newDefaultDocLoadParams(url)

	switch argt := arg.(type) {
	case runtime.Map:
		var input pageLoadParamsInput

		if err := sdk.Decode(argt, &input); err != nil {
			return PageLoadParams{}, err
		}

		if input.Driver != nil {
			res.Driver = *input.Driver
		}

		if input.Timeout != nil {
			res.Timeout = *input.Timeout
		}

		if input.UserAgent != nil {
			res.UserAgent = *input.UserAgent
		}

		if input.KeepCookies != nil {
			res.KeepCookies = *input.KeepCookies
		}

		if input.Headers != nil {
			res.Headers = input.Headers
		}

		if input.Viewport != nil {
			res.Viewport = input.Viewport
		}

		if input.Ignore != nil {
			res.Ignore = input.Ignore
		}

		if input.Charset != nil {
			res.Charset = *input.Charset
		}

		if input.Cookies != nil && input.Cookies != runtime.None {
			cookies, err := parseCookiesValue(ctx, input.Cookies)
			if err != nil {
				return PageLoadParams{}, err
			}

			res.Cookies = cookies
		}
	case runtime.String:
		res.Driver = arg.(runtime.String).String()
	}

	return res, nil
}
