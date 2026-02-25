package drivers

import "github.com/MontFerret/ferret/v2/pkg/runtime"

// Screenshot formats.
const (
	// ScreenshotFormatPNG represents the PNG format for screenshots.
	ScreenshotFormatPNG ScreenshotFormat = "png"

	// ScreenshotFormatJPEG represents the JPEG format for screenshots.
	ScreenshotFormatJPEG ScreenshotFormat = "jpeg"
)

type (
	// ScreenshotFormat represents the format of a screenshot.
	ScreenshotFormat string

	// ScreenshotParams defines parameters for the screenshot function.
	ScreenshotParams struct {
		X       runtime.Float
		Y       runtime.Float
		Width   runtime.Float
		Height  runtime.Float
		Format  ScreenshotFormat
		Quality runtime.Int
	}
)

func IsScreenshotFormatValid(format string) bool {
	value := ScreenshotFormat(format)

	return value == ScreenshotFormatPNG || value == ScreenshotFormatJPEG
}

func NewDefaultHTMLPDFParams() PDFParams {
	return PDFParams{
		Landscape:           runtime.False,
		DisplayHeaderFooter: runtime.False,
		PrintBackground:     runtime.False,
		Scale:               runtime.Float(1),
		PaperWidth:          runtime.Float(8.5),
		PaperHeight:         runtime.Float(11),
		MarginTop:           runtime.Float(0.4),
		MarginBottom:        runtime.Float(0.4),
		MarginLeft:          runtime.Float(0.4),
		MarginRight:         runtime.Float(0.4),
		PageRanges:          runtime.EmptyString,
		HeaderTemplate:      runtime.EmptyString,
		FooterTemplate:      runtime.EmptyString,
		PreferCSSPageSize:   runtime.False,
	}
}
