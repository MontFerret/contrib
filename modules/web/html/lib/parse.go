package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

type (
	ParseParams struct {
		Driver string `json:"driver"`
		drivers.ParseParams
	}

	parseParamsInput struct {
		Driver      string            `json:"driver"`
		Headers     runtime.Value     `json:"headers"`
		Cookies     runtime.Value     `json:"cookies"`
		Viewport    *drivers.Viewport `json:"viewport"`
		Content     []byte            `json:"content"`
		KeepCookies bool              `json:"keepCookies"`
	}
)

// Parse loads an HTML page from string or binary content.
//
// Options may select a driver, cookie reuse, cookies, headers, and viewport.
//
// @param html {String|Binary} HTML content.
// @param params {Object?} Driver and parsing options.
// @return {HTMLPage} Parsed page.
func Parse(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	if err := runtime.ValidateArgTypeAt(args, 0, runtime.TypeString, runtime.TypeBinary); err != nil {
		return runtime.None, err
	}

	var content []byte
	arg1 := args[0]

	switch v := arg1.(type) {
	case runtime.String:
		content = []byte(v)
	case runtime.Binary:
		content = v
	}

	var params ParseParams

	if len(args) > 1 {
		arg2, err := sdk.DecodeArg[runtime.Map](ctx, args, 1)

		if err != nil {
			return runtime.None, err
		}

		p, err := parseParseParams(ctx, content, arg2)

		if err != nil {
			return runtime.None, err
		}

		params = p
	} else {
		params = defaultParseParams(content)
	}

	drv, err := drivers.FromContext(ctx, params.Driver)

	if err != nil {
		return runtime.None, err
	}

	return drv.Parse(ctx, params.ParseParams)
}

func defaultParseParams(content []byte) ParseParams {
	return ParseParams{
		ParseParams: drivers.ParseParams{
			Content: content,
		},
		Driver: "",
	}
}

func parseParseParams(ctx context.Context, content []byte, arg runtime.Value) (ParseParams, error) {
	if err := runtime.AssertMap(arg); err != nil {
		return ParseParams{}, err
	}

	input := parseParamsInput{Content: content}

	if err := sdk.Decode(ctx, arg, &input, sdk.DisallowUnknownFields()); err != nil {
		return ParseParams{}, err
	}

	headers, err := parseHeadersValue(ctx, input.Headers)
	if err != nil {
		return ParseParams{}, err
	}

	res := ParseParams{
		Driver: input.Driver,
		ParseParams: drivers.ParseParams{
			Content:     input.Content,
			KeepCookies: input.KeepCookies,
			Viewport:    input.Viewport,
			Headers:     headers,
		},
	}

	if input.Cookies != nil && input.Cookies != runtime.None {
		cookies, err := parseCookiesValue(ctx, input.Cookies)
		if err != nil {
			return ParseParams{}, fmt.Errorf("cookies: %w", err)
		}

		res.Cookies = cookies
	}

	return res, nil
}
