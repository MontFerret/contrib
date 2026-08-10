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
		drivers.Params
		Driver  string        `json:"driver"`
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
		InitScript  *drivers.InitScript  `json:"initScript"`
	}
)

// Open loads an HTML page from a URL.
//
// Options may select a driver, timeout, user agent, cookie reuse, cookies,
// headers, ignored resources or status codes, viewport, and source charset.
//
// @param url {String} URL to load.
// @param params {Object?} Driver and page-loading options.
// @return {HTMLPage} Loaded page.
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

		if err := sdk.Decode(ctx, argt, &input, sdk.DisallowUnknownFields()); err != nil {
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

		if input.InitScript != nil {
			initScript, err := drivers.NormalizeInitScript(input.InitScript)
			if err != nil {
				return PageLoadParams{}, err
			}

			res.InitScript = initScript
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
