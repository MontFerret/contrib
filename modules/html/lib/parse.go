package html

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

type ParseParams struct {
	drivers.ParseParams
	Driver string `json:"driver"`
}

// PARSE loads an HTML page from a given string or byte array
// @param {String} html - HTML string to parse.
// @param {Object} [params] - An object containing the following properties:
// @param {String} [params.driver] - Name of a driver to parse with.
// @param {Boolean} [params.keepCookies=False] - Boolean value indicating whether to use cookies from previous sessions i.e. not to open a page in the Incognito mode.
// @param {HTTPCookies} [params.cookies] - Set of HTTP cookies to use during page loading.
// @param {HTTPHeaders} [params.headers] - Set of HTTP headers to use during page loading.
// @param {Object} [params.viewport] - Viewport params.
// @param {Int} [params.viewport.height] - Viewport height.
// @param {Int} [params.viewport.width] - Viewport width.
// @param {Float} [params.viewport.scaleFactor] - Viewport scale factor.
// @param {Boolean} [params.viewport.mobile] - Value that indicates whether to emulate mobile device.
// @param {Boolean} [params.viewport.landscape] - Value that indicates whether to render a page in landscape position.
// @return {HTMLPage} - Returns parsed and loaded HTML page.
func Parse(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 2); err != nil {
		return runtime.None, err
	}

	arg1 := args[0]

	if err := runtime.ValidateType(arg1, runtime.TypeString, runtime.TypeBinary); err != nil {
		return runtime.None, err
	}

	var content []byte

	switch v := arg1.(type) {
	case runtime.String:
		content = []byte(v)
	case runtime.Binary:
		content = v
	}

	var params ParseParams

	if len(args) > 1 {
		p, err := parseParseParams(content, args[1].(*runtime.Object))

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

func parseParseParams(content []byte, arg runtime.Value) (ParseParams, error) {
	if err := runtime.AssertMap(arg); err != nil {
		return ParseParams{}, err
	}

	res := defaultParseParams(content)

	if err := sdk.Decode(arg, &res); err != nil {
		return ParseParams{}, err
	}

	return res, nil
}
