package memory

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	cssxcommon "github.com/MontFerret/contrib/modules/web/html/drivers/internal/cssx"
)

func TestCSSXTraversalOps(t *testing.T) {
	root := mustSelection(t, `<ul><li>A</li><li>B</li><li>C</li><li>D</li></ul>`)
	items := cssxQueryAll(root, "li")

	cases := []struct {
		name string
		exp  cssxcommon.Expression
		args []any
		want []string
	}{
		{name: "first", exp: cssxcommon.ExpressionFirst, want: []string{"A"}},
		{name: "last", exp: cssxcommon.ExpressionLast, want: []string{"D"}},
		{name: "nth", exp: cssxcommon.ExpressionNth, args: []any{float64(1)}, want: []string{"B"}},
		{name: "take", exp: cssxcommon.ExpressionTake, args: []any{float64(2)}, want: []string{"A", "B"}},
		{name: "skip", exp: cssxcommon.ExpressionSkip, args: []any{float64(2)}, want: []string{"C", "D"}},
		{name: "slice", exp: cssxcommon.ExpressionSlice, args: []any{float64(1), float64(2)}, want: []string{"B", "C"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := cssxApplyCall(tc.exp, tc.args, []any{items}, nil)
			texts := nodeTexts(got)

			if len(texts) != len(tc.want) {
				t.Fatalf("expected %d values, got %d (%v)", len(tc.want), len(texts), texts)
			}

			for i := range tc.want {
				if texts[i] != tc.want[i] {
					t.Fatalf("expected %q at %d, got %q", tc.want[i], i, texts[i])
				}
			}
		})
	}
}

func TestCSSXWithin(t *testing.T) {
	root := mustSelection(t, `<div>
		<section class="product"><h2>P1</h2></section>
		<section class="product"><h2>P2</h2></section>
	</div>`)
	h2Nodes := cssxQueryAll(root, "h2")
	within := cssxApplyCall(
		cssxcommon.ExpressionWithin,
		[]any{".product"},
		[]any{h2Nodes},
		nil,
	)
	if texts := nodeTexts(within); len(texts) != 2 || texts[0] != "P1" || texts[1] != "P2" {
		t.Fatalf("unexpected :within result: %v", texts)
	}
}

func TestCSSXPredicateAndTransforms(t *testing.T) {
	root := mustSelection(t, `<div>
		<a href="/x"> Link </a>
		<a>Missing</a>
		<span class="price">$1,234.50</span>
		<time>2024-01-02</time>
	</div>`)
	links := cssxQueryAll(root, "a")
	first := cssxApplyCall(cssxcommon.ExpressionFirst, nil, []any{links}, nil)

	if got := cssxApplyCall(cssxcommon.ExpressionExists, nil, []any{links}, nil); got != true {
		t.Fatalf("expected exists=true, got %v", got)
	}

	if got := cssxApplyCall(cssxcommon.ExpressionEmpty, nil, []any{links}, nil); got != false {
		t.Fatalf("expected empty=false, got %v", got)
	}

	if got := cssxApplyCall(cssxcommon.ExpressionCount, nil, []any{links}, nil); got != 2 {
		t.Fatalf("expected count=2, got %v", got)
	}

	if got := firstValue(cssxApplyCall(cssxcommon.ExpressionText, nil, []any{first}, nil)); strings.TrimSpace(got.(string)) != "Link" {
		t.Fatalf("unexpected text result: %v", got)
	}

	attrs := cssxApplyCall(cssxcommon.ExpressionAttr, []any{"href"}, []any{links}, nil).([]any)
	if len(attrs) != 2 || attrs[0] != "/x" || attrs[1] != nil {
		t.Fatalf("unexpected attr result: %#v", attrs)
	}

	base, _ := url.Parse("https://example.com/base/")
	abs := firstValue(cssxApplyCall(cssxcommon.ExpressionURL, []any{"href"}, []any{first}, base))
	if abs != "https://example.com/x" {
		t.Fatalf("unexpected url result: %v", abs)
	}

	parsed := firstValue(cssxApplyCall(cssxcommon.ExpressionParseURL, nil, []any{abs}, base))
	obj, ok := parsed.(map[string]any)
	if !ok || obj["host"] != "example.com" {
		t.Fatalf("unexpected parseUrl result: %#v", parsed)
	}

	price := cssxApplyCall(cssxcommon.ExpressionFirst, nil, []any{cssxQueryAll(root, ".price")}, nil)
	if got := firstValue(cssxApplyCall(cssxcommon.ExpressionToNumber, nil, []any{price}, nil)); got != float64(1234.5) {
		t.Fatalf("unexpected toNumber result: %v", got)
	}

	day := cssxApplyCall(cssxcommon.ExpressionFirst, nil, []any{cssxQueryAll(root, "time")}, nil)
	if got := firstValue(cssxApplyCall(cssxcommon.ExpressionToDate, []any{"2006-01-02"}, []any{day}, nil)); got == nil {
		t.Fatalf("expected non-nil toDate result")
	}
}

func TestCSSXSelectionFilters(t *testing.T) {
	root := mustSelection(t, `<div>
		<a href="/a">same</a>
		<a href="/a">same</a>
		<a href="/b">other</a>
	</div>`)

	links := cssxQueryAll(root, "a")

	byAttr := cssxApplyCall(cssxcommon.ExpressionDedupeByAttr, []any{"href"}, []any{links}, nil)
	if texts := nodeTexts(byAttr); len(texts) != 2 || texts[0] != "same" || texts[1] != "other" {
		t.Fatalf("unexpected dedupeByAttr result: %v", texts)
	}

	byText := cssxApplyCall(cssxcommon.ExpressionDedupeByText, nil, []any{links}, nil)
	if texts := nodeTexts(byText); len(texts) != 2 || texts[0] != "same" || texts[1] != "other" {
		t.Fatalf("unexpected dedupeByText result: %v", texts)
	}

	filtered := cssxApplyCall(cssxcommon.ExpressionMatches, []any{`[href="/b"]`}, []any{links}, nil)
	if texts := nodeTexts(filtered); len(texts) != 1 || texts[0] != "other" {
		t.Fatalf("unexpected matches result: %v", texts)
	}
}

func TestCSSXSelectionModel(t *testing.T) {
	root := mustSelection(t, `<div>
		<section><p>A</p><p>B</p></section>
		<section><p>C</p></section>
		<a href="/a">A</a><a>B</a><a href="/a">A2</a>
	</div>`)

	paragraphs := cssxQueryAll(root, "p")
	texts := cssxApplyCall(cssxcommon.ExpressionText, nil, []any{paragraphs}, nil).([]any)
	if len(texts) != 3 || texts[0] != "A" || texts[1] != "B" || texts[2] != "C" {
		t.Fatalf("unexpected mapped texts: %#v", texts)
	}

	normalized := cssxApplyCall(cssxcommon.ExpressionNormalize, nil, []any{texts}, nil).([]any)
	if len(normalized) != 3 || normalized[2] != "C" {
		t.Fatalf("unexpected chained map: %#v", normalized)
	}

	attrs := cssxApplyCall(cssxcommon.ExpressionAttr, []any{"href"}, []any{cssxQueryAll(root, "a")}, nil).([]any)
	if len(attrs) != 3 || attrs[1] != nil {
		t.Fatalf("expected missing attribute placeholder: %#v", attrs)
	}

	compact := cssxApplyCall(cssxcommon.ExpressionCompact, nil, []any{attrs}, nil).([]any)
	if len(compact) != 2 {
		t.Fatalf("expected compacted values, got %#v", compact)
	}

	distinct := cssxApplyCall(cssxcommon.ExpressionDistinct, nil, []any{compact}, nil).([]any)
	if len(distinct) != 1 || distinct[0] != "/a" {
		t.Fatalf("expected distinct values, got %#v", distinct)
	}

	shared := map[string]any{"value": "same"}
	objects := cssxApplyCall(
		cssxcommon.ExpressionDistinct,
		nil,
		[]any{[]any{shared, shared, map[string]any{"value": "same"}}},
		nil,
	).([]any)
	if len(objects) != 2 {
		t.Fatalf("expected object identity deduplication, got %#v", objects)
	}

	parents := cssxApplyCall(cssxcommon.ExpressionParent, nil, []any{paragraphs}, nil).([]any)
	if len(parents) != 3 || parents[0] != parents[1] {
		t.Fatalf("expected duplicate-preserving traversal, got %#v", parents)
	}

	hasParagraph := cssxApplyCall(cssxcommon.ExpressionHas, []any{"p"}, []any{parents}, nil).([]any)
	if len(hasParagraph) != 3 {
		t.Fatalf("expected duplicate-preserving predicate result, got %#v", hasParagraph)
	}

	notSection := cssxApplyCall(cssxcommon.ExpressionNot, []any{"section"}, []any{parents}, nil).([]any)
	if len(notSection) != 0 {
		t.Fatalf("expected :not to remove matching nodes, got %#v", notSection)
	}

	invalid := cssxApplyCall(cssxcommon.ExpressionMatches, []any{"["}, []any{paragraphs}, nil).([]any)
	if len(invalid) != 0 {
		t.Fatalf("expected invalid relative selector to return empty selection, got %#v", invalid)
	}

	invalidNot := cssxApplyCall(cssxcommon.ExpressionNot, []any{"["}, []any{paragraphs}, nil).([]any)
	if len(invalidNot) != 0 {
		t.Fatalf("expected invalid negated selector to return empty selection, got %#v", invalidNot)
	}

	if got := cssxApplyCall(cssxcommon.ExpressionOne, nil, []any{[]any{nil}}, nil); got != true {
		t.Fatalf("expected one NONE slot to satisfy :one, got %v", got)
	}

	if got := cssxApplyCall(cssxcommon.ExpressionCount, nil, []any{[]any{nil}}, nil); got != 1 {
		t.Fatalf("expected NONE slot to count, got %v", got)
	}
}

func mustDocument(t *testing.T, input string) *goquery.Document {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	return doc
}

func mustSelection(t *testing.T, input string) *goquery.Selection {
	t.Helper()

	return mustDocument(t, input).Selection
}

func nodeTexts(value any) []string {
	nodes := cssxToNodes(value)
	out := make([]string, 0, len(nodes))

	for _, node := range nodes {
		out = append(out, strings.TrimSpace(cssxTextContent(node)))
	}

	return out
}

func firstValue(value any) any {
	items := cssxToArray(value)
	if len(items) == 0 {
		return nil
	}

	return items[0]
}
