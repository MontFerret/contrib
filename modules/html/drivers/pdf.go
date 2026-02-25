package drivers

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// PDFParams represents the arguments for PrintToPDF function.
type PDFParams struct {
	// Paper orientation. Defaults to false.
	Landscape runtime.Boolean `json:"landscape"`
	// Display Data and footer. Defaults to false.
	DisplayHeaderFooter runtime.Boolean `json:"displayHeaderFooter"`
	// Print background graphics. Defaults to false.
	PrintBackground runtime.Boolean `json:"printBackground"`
	// Scale of the webpage rendering. Defaults to 1.
	Scale runtime.Float `json:"scale"`
	// Paper width in inches. Defaults to 8.5 inches.
	PaperWidth runtime.Float `json:"paperWidth"`
	// Paper height in inches. Defaults to 11 inches.
	PaperHeight runtime.Float `json:"paperHeight"`
	// Top margin in inches. Defaults to 1cm (~0.4 inches).
	MarginTop runtime.Float `json:"marginTop"`
	// Bottom margin in inches. Defaults to 1cm (~0.4 inches).
	MarginBottom runtime.Float `json:"marginBottom"`
	// Left margin in inches. Defaults to 1cm (~0.4 inches).
	MarginLeft runtime.Float `json:"marginLeft"`
	// Right margin in inches. Defaults to 1cm (~0.4 inches).
	MarginRight runtime.Float `json:"marginRight"`
	// Paper ranges to print, e.g., '1-5, 8, 11-13'. Defaults to the empty string, which means print all pages.
	PageRanges runtime.String `json:"pageRanges"`
	// HTML template for the print runtime. Should be valid HTML markup with following classes used to inject printing Data into them: - `date`: formatted print date - `title`: document title - `url`: document location - `pageNumber`: current page number - `totalPages`: total pages in the document
	// For example, `<span class=title></span>` would generate span containing the title.
	HeaderTemplate runtime.String `json:"headerTemplate"`
	// HTML template for the print footer. Should use the same format as the `headerTemplate`.
	FooterTemplate runtime.String `json:"footerTemplate"`
	// Whether or not to prefer page size as defined by css.
	// Defaults to false, in which case the content will be scaled to fit the paper size.
	PreferCSSPageSize runtime.Boolean `json:"preferCSSPageSize"`
}
