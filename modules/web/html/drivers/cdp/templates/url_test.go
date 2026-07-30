package templates

import (
	"strings"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestDocumentURLTemplatesUseLiveDocumentState(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		fn   string
		want string
	}{
		{name: "current URL", fn: GetDocumentURL().String(), want: "document.URL"},
		{name: "base URL", fn: GetBaseURL().String(), want: "document.baseURI"},
		{name: "resolve URL", fn: ResolveURL(runtime.NewString("asset.png")).String(), want: "new URL(url, document.baseURI).href"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !strings.Contains(tc.fn, tc.want) {
				t.Fatalf("expected template to contain %q, got %q", tc.want, tc.fn)
			}
		})
	}

	if got := ResolveURL(runtime.NewString("asset.png")).Length(); got != 1 {
		t.Fatalf("expected resolve template to receive one argument, got %d", got)
	}
}
