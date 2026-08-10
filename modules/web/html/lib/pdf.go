package lib

import (
	"context"
	"regexp"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func ValidatePageRanges(pageRanges string) (bool, error) {
	match, err := regexp.Match(`^(([1-9][0-9]*|[1-9][0-9]*)(\s*-\s*|\s*,\s*|))*$`, []byte(pageRanges))

	if err != nil {
		return false, err
	}

	return match, nil
}

// PDF renders a page or URL as a PDF document.
//
// Options control orientation, headers and footers, background graphics, scale,
// paper size, margins, page ranges, templates, and CSS page-size preference.
//
// @param target {HTMLPage|String} Page or URL to render.
// @param params {Object?} PDF layout and rendering options.
// @return {Binary} Rendered PDF bytes.
func PDF(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	err := runtime.ValidateArgs(args, 1, 2)

	if err != nil {
		return runtime.None, err
	}

	arg1 := args[0]
	page, closeAfter, err := OpenOrCastPage(ctx, arg1)

	if err != nil {
		return runtime.None, err
	}

	defer func() {
		if closeAfter {
			page.Close()
		}
	}()

	var pdfParams drivers.PDFParams

	if len(args) == 2 {
		paramArg, err := runtime.CastMap(args[1])

		if err != nil {
			return runtime.None, err
		}

		param, err := parsePDFParams(ctx, paramArg)

		if err != nil {
			return runtime.None, err
		}

		pdfParams = param
	}

	target, err := drivers.ToPageSnapshotTarget(page)
	if err != nil {
		return runtime.None, err
	}

	pdf, err := target.PrintToPDF(ctx, pdfParams)

	if err != nil {
		return runtime.None, err
	}

	return pdf, nil
}

func parsePDFParams(ctx context.Context, values runtime.Map) (drivers.PDFParams, error) {
	var pdfParams drivers.PDFParams

	if err := sdk.Decode(ctx, values, &pdfParams, sdk.DisallowUnknownFields()); err != nil {
		return drivers.PDFParams{}, err
	}

	return pdfParams, nil
}
