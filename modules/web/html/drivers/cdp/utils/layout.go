package utils

import "github.com/mafredri/cdp/protocol/page"

func GetLayoutViewportWH(metrics *page.GetLayoutMetricsReply) (width int, height int) {
	if metrics.CSSLayoutViewport.ClientWidth > 0 {
		width = metrics.CSSLayoutViewport.ClientWidth
	} else {
		// Chrome version <=89
		//lint:ignore SA1019 Older Chrome versions may only populate the deprecated layout viewport fields.
		width = metrics.LayoutViewport.ClientWidth
	}

	if metrics.CSSLayoutViewport.ClientHeight > 0 {
		height = metrics.CSSLayoutViewport.ClientHeight
	} else {
		// Chrome version <=89
		//lint:ignore SA1019 Older Chrome versions may only populate the deprecated layout viewport fields.
		height = metrics.LayoutViewport.ClientHeight
	}

	return
}
