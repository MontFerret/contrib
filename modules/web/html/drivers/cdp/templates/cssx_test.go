package templates

import (
	"strings"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestCSSXCompilesAllSupportedSelectors(t *testing.T) {
	cases := []struct {
		name string
		exp  string
	}{
		{name: "first", exp: `:first(section)`},
		{name: "last", exp: `:last(section)`},
		{name: "nth", exp: `:nth(0, section)`},
		{name: "take", exp: `:take(2, section)`},
		{name: "skip", exp: `:skip(1, section)`},
		{name: "slice", exp: `:slice(1, 2, section)`},
		{name: "within", exp: `:within(section, :text(:first(h1)))`},
		{name: "parent", exp: `:parent(:first(section))`},
		{name: "closest", exp: `:closest(section, :first(h1))`},
		{name: "children", exp: `:children(li, :first(ul))`},
		{name: "next", exp: `:next(div, :first(span))`},
		{name: "prev", exp: `:prev(div, :first(span))`},
		{name: "exists", exp: `:exists(section)`},
		{name: "empty", exp: `:empty(section)`},
		{name: "has", exp: `:has(div, :first(section))`},
		{name: "matches", exp: `:matches(section, :first(section))`},
		{name: "count", exp: `:count(section)`},
		{name: "indexOf", exp: `:indexOf(section, :first(section))`},
		{name: "len", exp: `:len(:texts(section))`},
		{name: "text", exp: `:text(:first(section))`},
		{name: "texts", exp: `:texts(section)`},
		{name: "ownText", exp: `:ownText(section)`},
		{name: "normalize", exp: `:normalize(:text(section))`},
		{name: "trim", exp: `:trim(:text(section))`},
		{name: "join", exp: `:join(", ", :texts(section))`},
		{name: "attr", exp: `:attr("href", :first(a))`},
		{name: "attrs", exp: `:attrs("href", a)`},
		{name: "prop", exp: `:prop("value", :first(input))`},
		{name: "html", exp: `:html(:first(section))`},
		{name: "outerHtml", exp: `:outerHtml(:first(section))`},
		{name: "value", exp: `:value(:first(input))`},
		{name: "absUrl", exp: `:absUrl(:attr("href", :first(a)))`},
		{name: "url", exp: `:url("href", :first(a))`},
		{name: "parseUrl", exp: `:parseUrl(:url("href", :first(a)))`},
		{name: "filter", exp: `:filter(section, section)`},
		{name: "withAttr", exp: `:withAttr("href", a)`},
		{name: "withText", exp: `:withText("Sale", section)`},
		{name: "dedupeByAttr", exp: `:dedupeByAttr("href", a)`},
		{name: "dedupeByText", exp: `:dedupeByText(section)`},
		{name: "replace", exp: `:replace("\\s+", " ", :text(section))`},
		{name: "regex", exp: `:regex("(\\d+)", 1, :text(section))`},
		{name: "toNumber", exp: `:toNumber(:text(section))`},
		{name: "toDate", exp: `:toDate("2006-01-02", :text(time))`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fn, err := CSSX(cdpruntime.RemoteObjectID("obj"), runtime.NewString(tc.exp))

			if err != nil {
				t.Fatalf("expected expression to compile, got %v", err)
			}

			if fn == nil {
				t.Fatalf("expected non-nil function")
			}

			js := fn.String()
			if !strings.Contains(js, "const ops =") {
				t.Fatalf("expected generated JS to include ops descriptor")
			}
		})
	}
}

func TestCSSXUsesCallNameField(t *testing.T) {
	fn, err := CSSX(cdpruntime.RemoteObjectID("obj"), runtime.NewString(`:first(div)`))

	if err != nil {
		t.Fatalf("expected expression to compile, got %v", err)
	}

	if !strings.Contains(fn.String(), `":first"`) {
		t.Fatalf("expected generated JS to include normalized call name, got %s", fn.String())
	}
}

func TestCSSXRejectsInvalidArgs(t *testing.T) {
	cases := []string{
		`:nth("1", section)`,
		`:slice(1, section)`,
		`:attr(1, a)`,
		`:replace("\\s+", :text(section))`,
		`:regex(1, :text(section))`,
	}

	for _, exp := range cases {
		exp := exp
		t.Run(exp, func(t *testing.T) {
			_, err := CSSX(cdpruntime.RemoteObjectID("obj"), runtime.NewString(exp))

			if err == nil {
				t.Fatalf("expected validation error for %s", exp)
			}
		})
	}
}

func TestCSSXSliceOffsetLimitSemantics(t *testing.T) {
	fn, err := CSSX(cdpruntime.RemoteObjectID("obj"), runtime.NewString(`:slice(10, 5, section)`))

	if err != nil {
		t.Fatalf("expected expression to compile, got %v", err)
	}

	js := fn.String()

	if !strings.Contains(js, "start + count") {
		t.Fatalf("expected offset+limit slice semantics, got %s", js)
	}

	if !strings.Contains(js, `":slice"`) {
		t.Fatalf("expected slice op to be encoded in JS, got %s", js)
	}
}

func TestCSSXCompilesNestedMultiArity(t *testing.T) {
	fn, err := CSSX(cdpruntime.RemoteObjectID("obj"), runtime.NewString(`:indexOf(.item, :first(.selected))`))

	if err != nil {
		t.Fatalf("expected nested expression to compile, got %v", err)
	}

	js := fn.String()

	if !strings.Contains(js, `":indexOf"`) {
		t.Fatalf("expected indexOf operation in generated JS, got %s", js)
	}

	if !strings.Contains(js, `"arity":2`) {
		t.Fatalf("expected indexOf arity 2 in generated JS, got %s", js)
	}
}
