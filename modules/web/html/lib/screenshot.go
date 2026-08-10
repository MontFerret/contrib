package lib

import (
	"context"
	"fmt"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// Screenshot captures a screenshot of a page or URL.
//
// The format is JPEG or PNG; quality applies only to JPEG.
//
// @param target {HTMLPage|String} Page or URL to capture.
// @param params {Object?} Capture bounds, format, and quality options.
// @return {Binary} Screenshot bytes.
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

		parsed, err := parseScreenshotParams(ctx, values)

		if err != nil {
			return runtime.None, err
		}

		screenshotParams = parsed
	} else {
		screenshotParams = defaultScreenshotParams()
	}

	target, err := drivers.ToPageSnapshotTarget(page)
	if err != nil {
		return runtime.None, err
	}

	scr, err := target.CaptureScreenshot(ctx, screenshotParams)

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

func parseScreenshotParams(ctx context.Context, values runtime.Map) (drivers.ScreenshotParams, error) {
	res := defaultScreenshotParams()

	if err := sdk.Decode(ctx, values, &res, sdk.DisallowUnknownFields()); err != nil {
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
