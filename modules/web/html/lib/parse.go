package lib

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

type ParseParams struct {
	Driver string `json:"driver"`
	drivers.ParseParams
}

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

	res := defaultParseParams(content)

	if err := sdk.Decode(ctx, arg, &res, sdk.DisallowUnknownFields()); err != nil {
		return ParseParams{}, err
	}

	return res, nil
}
