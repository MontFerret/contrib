package html

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// SCREENSHOT takes a screenshot of a given page.
// @param {HTMLPage|String} target - Target page or url.
// @param {Object} [params] - An object containing the following properties :
// @param {Float | Int} [params.x=0] - X position of the viewport.
// @param {Float | Int} [params.y=0] - Y position of the viewport.
// @param {Float | Int} [params.width] - Width of the viewport.
// @param {Float | Int} [params.height] - Height of the viewport.
// @param {String} [params.format="jpeg"] - Either "jpeg" or "png".
// @param {Int} [params.quality=100] - Quality, in [0, 100], only for jpeg format.
// @return {Binary} - Screenshot in binary format.
func Screenshot(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	arg1 := args[0]

	err = runtime.ValidateType(arg1, drivers.HTMLPageType, runtime.TypeString)

	if err != nil {
		return runtime.None, err
	}

	page, closeAfter, err := OpenOrCastPage(ctx, arg1)

	if err != nil {
		return runtime.None, err
	}

	defer func() {
		if closeAfter {
			page.Close()
		}
	}()

	var screenshotParams drivers.ScreenshotParams

	if len(args) == 2 {
		values, err := runtime.CastMap(args[1])

		if err != nil {
			return runtime.None, err
		}

		parsed, err := parseScreenshotParams(values)

		if err != nil {
			return runtime.None, err
		}

		screenshotParams = parsed
	} else {
		screenshotParams = defaultScreenshotParams()
	}

	scr, err := page.CaptureScreenshot(ctx, screenshotParams)

	if err != nil {
		return runtime.None, err
	}

	return scr, nil
}

func defaultScreenshotParams() drivers.ScreenshotParams {
	return drivers.ScreenshotParams{
		X:       0,
		Y:       0,
		Width:   -1,
		Height:  -1,
		Format:  drivers.ScreenshotFormatJPEG,
		Quality: 100,
	}
}

func parseScreenshotParams(values runtime.Map) (drivers.ScreenshotParams, error) {
	res := defaultScreenshotParams()

	if err := sdk.Decode(values, &res); err != nil {
		return drivers.ScreenshotParams{}, err
	}

	if res.Format != drivers.ScreenshotFormatJPEG && res.Format != drivers.ScreenshotFormatPNG {
		return drivers.ScreenshotParams{}, fmt.Errorf("unsupported format: %s", res.Format)
	}

	if res.Quality < 0 || res.Quality > 100 {
		return drivers.ScreenshotParams{}, fmt.Errorf("quality should be in [0, 100], got %d", res.Quality)
	}

	return res, nil
}
