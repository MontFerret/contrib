package utils

import (
	"testing"

	"github.com/mafredri/cdp/protocol/page"
)

func TestGetLayoutViewportWHUsesCSSLayoutViewport(t *testing.T) {
	metrics := &page.GetLayoutMetricsReply{
		CSSLayoutViewport: page.LayoutViewport{
			ClientWidth:  800,
			ClientHeight: 600,
		},
		LayoutViewport: page.LayoutViewport{
			ClientWidth:  1024,
			ClientHeight: 768,
		},
	}

	width, height := GetLayoutViewportWH(metrics)

	if width != 800 || height != 600 {
		t.Fatalf("expected CSS layout viewport dimensions, got %dx%d", width, height)
	}
}

func TestGetLayoutViewportWHFallsBackToLegacyLayoutViewport(t *testing.T) {
	metrics := &page.GetLayoutMetricsReply{
		LayoutViewport: page.LayoutViewport{
			ClientWidth:  1024,
			ClientHeight: 768,
		},
	}

	width, height := GetLayoutViewportWH(metrics)

	if width != 1024 || height != 768 {
		t.Fatalf("expected legacy layout viewport dimensions, got %dx%d", width, height)
	}
}
