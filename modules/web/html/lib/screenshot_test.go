package lib

import (
	"strings"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestScreenshotOptions(t *testing.T) {
	defaults := drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, Width: -1, Height: -1, Quality: 100}

	for _, tc := range []struct {
		name    string
		options map[string]runtime.Value
		want    drivers.ScreenshotParams
	}{
		{"omitted", nil, defaults},
		{"empty", map[string]runtime.Value{}, defaults},
		{"format", map[string]runtime.Value{"format": runtime.String("png")},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatPNG, Width: -1, Height: -1, Quality: 100}},
		{"x", map[string]runtime.Value{"x": runtime.Float(1.5)},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, X: 1.5, Width: -1, Height: -1, Quality: 100}},
		{"y", map[string]runtime.Value{"y": runtime.Float(2.25)},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, Y: 2.25, Width: -1, Height: -1, Quality: 100}},
		{"width", map[string]runtime.Value{"width": runtime.Int(64)},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, Width: 64, Height: -1, Quality: 100}},
		{"height", map[string]runtime.Value{"height": runtime.Int(32)},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, Width: -1, Height: 32, Quality: 100}},
		{"quality", map[string]runtime.Value{"quality": runtime.Int(75)},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, Width: -1, Height: -1, Quality: 75}},
		{"minimum quality", map[string]runtime.Value{"quality": runtime.Int(0)},
			drivers.ScreenshotParams{Format: drivers.ScreenshotFormatJPEG, Width: -1, Height: -1, Quality: 0}},
		{"maximum quality", map[string]runtime.Value{"quality": runtime.Int(100)}, defaults},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := newTestPage(t, "<html></html>")
			args := []runtime.Value{page}

			if tc.options != nil {
				args = append(args, runtime.NewObjectWith(tc.options))
			}

			_, err := Screenshot(t.Context(), args...)
			if err != nil {
				t.Fatal(err)
			}

			if page.screenshotParams == nil {
				t.Fatal("CaptureScreenshot was not called")
			}

			if got := *page.screenshotParams; got != tc.want {
				t.Errorf("CaptureScreenshot received %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestScreenshotRejectsInvalidOptions(t *testing.T) {
	for _, tc := range []struct {
		key     string
		value   runtime.Value
		wantErr string
	}{
		{"format", runtime.String("gif"), "unsupported format"},
		{"quality", runtime.Int(-1), "quality should be in [0, 100]"},
		{"quality", runtime.Int(101), "quality should be in [0, 100]"},
		{"width", runtime.String("64"), "$.width: invalid argument type"},
		{"extra", runtime.True, `unknown field "extra"`},
	} {
		t.Run(tc.key+"="+tc.value.String(), func(t *testing.T) {
			page := newTestPage(t, "<html></html>")
			options := runtime.NewObjectWith(map[string]runtime.Value{tc.key: tc.value})

			_, err := Screenshot(t.Context(), page, options)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("expected an error containing %q, got %v", tc.wantErr, err)
			}

			if page.screenshotParams != nil {
				t.Error("CaptureScreenshot was called with invalid options")
			}
		})
	}
}
