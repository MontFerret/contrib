package templates

import (
	"strings"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
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
		{name: "within", exp: `:within(".card", h1)`},
		{name: "parent", exp: `:parent(:first(section))`},
		{name: "closest", exp: `:closest("section", h1)`},
		{name: "children", exp: `:children("li", ul)`},
		{name: "next", exp: `:next("div", span)`},
		{name: "prev", exp: `:prev("div", span)`},
		{name: "siblings", exp: `:siblings("li", li.active)`},
		{name: "exists", exp: `:exists(section)`},
		{name: "empty", exp: `:empty(section)`},
		{name: "has", exp: `:has("h1", section)`},
		{name: "matches", exp: `:matches(".active", section)`},
		{name: "not", exp: `:not(".active", section)`},
		{name: "count", exp: `:count(section)`},
		{name: "one", exp: `:one(section)`},
		{name: "indexOf", exp: `:indexOf(section, :first(section))`},
		{name: "len", exp: `:len(:text(section))`},
		{name: "text", exp: `:text(:first(section))`},
		{name: "ownText", exp: `:ownText(section)`},
		{name: "normalize", exp: `:normalize(:text(section))`},
		{name: "trim", exp: `:trim(:text(section))`},
		{name: "join", exp: `:join(", ", :text(section))`},
		{name: "attr", exp: `:attr("href", :first(a))`},
		{name: "prop", exp: `:prop("value", :first(input))`},
		{name: "html", exp: `:html(:first(section))`},
		{name: "outerHtml", exp: `:outerHtml(:first(section))`},
		{name: "value", exp: `:value(:first(input))`},
		{name: "absUrl", exp: `:absUrl(:attr("href", :first(a)))`},
		{name: "url", exp: `:url("href", :first(a))`},
		{name: "parseUrl", exp: `:parseUrl(:url("href", :first(a)))`},
		{name: "compact", exp: `:compact(:attr("href", a))`},
		{name: "distinct", exp: `:distinct(:text(section))`},
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

	if !strings.Contains(fn.String(), `"family":"cardinality"`) {
		t.Fatalf("expected generated JS to include operation family, got %s", fn.String())
	}
}

func TestCSSXModifierTemplatesUseDistinctFinalizers(t *testing.T) {
	cases := []struct {
		build func(cdpruntime.RemoteObjectID, runtime.String) (*eval.Function, error)
		name  string
		want  string
	}{
		{
			name: "list",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSX(id, exp)
			},
			want: "return [result];",
		},
		{
			name: "one",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSXOne(id, exp)
			},
			want: "return result.length > 0 ? result[0] : null;",
		},
		{
			name: "count",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSXCount(id, exp)
			},
			want: "return 1;",
		},
		{
			name: "exists",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSXExists(id, exp)
			},
			want: "return result != null;",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fn, err := tc.build(cdpruntime.RemoteObjectID("obj"), runtime.NewString(`:value(:first(input))`))
			if err != nil {
				t.Fatalf("expected expression to compile, got %v", err)
			}

			js := fn.String()
			if !strings.Contains(js, "const ops =") {
				t.Fatalf("expected state machine JS, got %s", js)
			}

			if !strings.Contains(js, tc.want) {
				t.Fatalf("expected generated JS to include %q, got %s", tc.want, js)
			}
		})
	}
}

func TestCSSXModifierTemplatesUseSimpleSelectorFastPaths(t *testing.T) {
	cases := []struct {
		build func(cdpruntime.RemoteObjectID, runtime.String) (*eval.Function, error)
		name  string
		want  string
	}{
		{
			name: "one",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSXOne(id, exp)
			},
			want: "return el.querySelector(selector);",
		},
		{
			name: "count",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSXCount(id, exp)
			},
			want: "return el.querySelectorAll(selector).length;",
		},
		{
			name: "exists",
			build: func(id cdpruntime.RemoteObjectID, exp runtime.String) (*eval.Function, error) {
				return CSSXExists(id, exp)
			},
			want: "return el.querySelector(selector) != null;",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			fn, err := tc.build(cdpruntime.RemoteObjectID("obj"), runtime.NewString(`.track[data-index='0']`))
			if err != nil {
				t.Fatalf("expected expression to compile, got %v", err)
			}

			js := fn.String()
			if strings.Contains(js, "const ops =") {
				t.Fatalf("expected simple selector fast path, got %s", js)
			}

			if !strings.Contains(js, tc.want) {
				t.Fatalf("expected generated JS to include %q, got %s", tc.want, js)
			}
		})
	}
}

func TestCSSXRejectsInvalidArgs(t *testing.T) {
	cases := []string{
		`:nth("1", section)`,
		`:slice(1, section)`,
		`:attr(1, a)`,
		`:texts(a)`,
		`:attrs("href", a)`,
		`:filter(a, a)`,
		`:has(.price, .product)`,
		`:closest(.card)`,
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
