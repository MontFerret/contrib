package templates

import (
	"strings"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestXPathModifierTemplatesUseShapeSpecificResults(t *testing.T) {
	cases := []struct {
		build func(cdpruntime.RemoteObjectID, runtime.String) string
		name  string
		want  []string
	}{
		{
			name: "one",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) string {
				return XPathOne(id, exp).String()
			},
			want: []string{
				"const item = out.iterateNext();",
				"return out.numberValue;",
				"return node != null ? unwrap(node) : null;",
			},
		},
		{
			name: "count",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) string {
				return XPathCount(id, exp).String()
			},
			want: []string{
				"let count = 0;",
				"return out.snapshotLength;",
				"return 1;",
			},
		},
		{
			name: "exists",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) string {
				return XPathExists(id, exp).String()
			},
			want: []string{
				"return out.iterateNext() != null;",
				"return out.snapshotLength > 0;",
				"return true;",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			js := tc.build(cdpruntime.RemoteObjectID("obj"), runtime.NewString(`//article`))

			for _, want := range tc.want {
				if !strings.Contains(js, want) {
					t.Fatalf("expected generated JS to include %q, got %s", want, js)
				}
			}
		})
	}
}
