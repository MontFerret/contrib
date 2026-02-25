package html

import (
	"context"
	"regexp"

	"github.com/MontFerret/contrib/modules/html/drivers"
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

// PDF prints a PDF of the current page.
// @param {HTMLPage | String}target - Target page or url.
// @param {Object} [params] - An object containing the following properties:
// @param {Bool} [params.landscape=False] - Paper orientation.
// @param {Bool} [params.displayHeaderFooter=False] - Display header and footer.
// @param {Bool} [params.printBackground=False] - Print background graphics.
// @param {Float} [params.scale=1] - Scale of the webpage rendering.
// @param {Float} [params.paperWidth=22] - Paper width in inches.
// @param {Float} [params.paperHeight=28] - Paper height in inches.
// @param {Float} [params.marginTo=1] - Top margin in inches.
// @param {Float} [params.marginBottom=1] - Bottom margin in inches.
// @param {Float} [params.marginLeft=1] - Left margin in inches.
// @param {Float} [params.marginRight=1] - Right margin in inches.
// @param {String} [params.pageRanges] - Paper ranges to print, e.g., '1-5, 8, 11-13'.
// @param {String} [params.headerTemplate] - HTML template for the print header. Should be valid HTML markup with following classes used to inject printing values into them: - `date`: formatted print date - `title`: document title - `url`: document location - `pageNumber`: current page number - `totalPages`: total pages in the document For example, `<span class=title></span>` would generate span containing the title.
// @param {String} [params.footerTemplate] - HTML template for the print footer. Should use the same format as the `headerTemplate`.
// @param {Bool} [params.preferCSSPageSize=False] - Whether or not to prefer page size as defined by css. Defaults to false, in which case the content will be scaled to fit the paper size. *
// @return {Binary} - PDF document in binary format.
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

		param, err := parsePDFParams(paramArg)

		if err != nil {
			return runtime.None, err
		}

		pdfParams = param
	}

	pdf, err := page.PrintToPDF(ctx, pdfParams)

	if err != nil {
		return runtime.None, err
	}

	return pdf, nil
}

func parsePDFParams(values runtime.Map) (drivers.PDFParams, error) {
	var pdfParams drivers.PDFParams

	if err := sdk.Decode(values, &pdfParams); err != nil {
		return drivers.PDFParams{}, err
	}

	return pdfParams, nil
}
