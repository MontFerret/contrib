package drivers

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// PDFParams represents the arguments for PrintToPDF function.
type PDFParams struct {
	PageRanges          runtime.String  `json:"pageRanges"`
	FooterTemplate      runtime.String  `json:"footerTemplate"`
	HeaderTemplate      runtime.String  `json:"headerTemplate"`
	MarginTop           runtime.Float   `json:"marginTop"`
	PaperWidth          runtime.Float   `json:"paperWidth"`
	PaperHeight         runtime.Float   `json:"paperHeight"`
	MarginBottom        runtime.Float   `json:"marginBottom"`
	MarginLeft          runtime.Float   `json:"marginLeft"`
	MarginRight         runtime.Float   `json:"marginRight"`
	Scale               runtime.Float   `json:"scale"`
	Landscape           runtime.Boolean `json:"landscape"`
	PrintBackground     runtime.Boolean `json:"printBackground"`
	DisplayHeaderFooter runtime.Boolean `json:"displayHeaderFooter"`
	PreferCSSPageSize   runtime.Boolean `json:"preferCSSPageSize"`
}
