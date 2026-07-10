package templates

import (
	"strings"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestGetDOMPropertyTemplateNormalizesNativeResults(t *testing.T) {
	fn := GetDOMProperty(cdpruntime.RemoteObjectID("node"), runtime.NewString("selectedOptions"))

	if fn.Length() != 2 {
		t.Fatalf("expected element ref and property name args, got %d", fn.Length())
	}

	js := fn.String()
	for _, want := range []string{
		"const value = el[name];",
		"return value;",
		`typeof value === "function"`,
		"value instanceof Node",
		"Array.isArray(value)",
		"value instanceof NodeList",
		"value instanceof HTMLCollection",
		"return Array.from(value);",
		"return undefined;",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("expected generated JS to include %q, got %s", want, js)
		}
	}
}
