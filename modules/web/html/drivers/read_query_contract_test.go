package drivers_test

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const readQueryContractMarkup = `<html><head><title>Contract Fixture</title></head><body><section id="root"><article id="hero" class="card primary" data-role="hero" style="color:red;display:block"><h1>Ferret</h1><p class="lead">Browser automation</p><a class="action" href="/next" title="Continue">Next</a><input id="search" value="find me"/></article><article id="secondary" class="card secondary" data-role="related"><h2>Other</h2></article></section></body></html>`

func TestReadQueryContractForDocumentAndElement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		target func(*testing.T) drivers.QueryTarget
	}{
		{
			name: "document",
			target: func(t *testing.T) drivers.QueryTarget {
				return newMemoryDocument(t, readQueryContractMarkup)
			},
		},
		{
			name: "element",
			target: func(t *testing.T) drivers.QueryTarget {
				return newMemoryDocument(t, readQueryContractMarkup).GetElement()
			},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			target := tt.target(t)

			assertSharedQuerySemantics(t, ctx, target)
			assertSharedReadSemantics(t, ctx, target)
		})
	}
}

func TestReadQueryContractForPageQueryResolver(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	page := newMemoryPage(t, readQueryContractMarkup)

	target, err := drivers.ToQueryTarget(page)
	if err != nil {
		t.Fatalf("expected page query resolver: %v", err)
	}

	assertSharedQuerySemantics(t, ctx, target)
}

func assertSharedQuerySemantics(t *testing.T, ctx context.Context, target drivers.QueryTarget) {
	t.Helper()

	found, err := target.QuerySelector(ctx, drivers.NewCSSSelector("#hero"))
	if err != nil {
		t.Fatalf("query selector hit failed: %v", err)
	}

	hero := mustElementFromValue(t, found)
	assertElementID(t, ctx, hero, "hero")

	missing, err := target.QuerySelector(ctx, drivers.NewCSSSelector("#missing"))
	if err != nil {
		t.Fatalf("query selector miss failed: %v", err)
	}

	if missing != runtime.None {
		t.Fatalf("expected query selector miss to return None, got %T", missing)
	}

	all, err := target.QuerySelectorAll(ctx, drivers.NewCSSSelector(".card"))
	if err != nil {
		t.Fatalf("query selector all failed: %v", err)
	}

	assertElementIDs(t, ctx, all, []string{"hero", "secondary"})

	count, err := target.CountBySelector(ctx, drivers.NewCSSSelector(".card"))
	if err != nil {
		t.Fatalf("count by selector failed: %v", err)
	}

	if runtime.CompareValues(count, runtime.NewInt(2)) != 0 {
		t.Fatalf("expected selector count 2, got %v", count)
	}

	exists, err := target.ExistsBySelector(ctx, drivers.NewCSSSelector(".card"))
	if err != nil {
		t.Fatalf("exists by selector failed: %v", err)
	}

	if exists != runtime.True {
		t.Fatalf("expected selector to exist, got %v", exists)
	}

	missingExists, err := target.ExistsBySelector(ctx, drivers.NewCSSSelector(".missing"))
	if err != nil {
		t.Fatalf("missing exists by selector failed: %v", err)
	}

	if missingExists != runtime.False {
		t.Fatalf("expected missing selector to be absent, got %v", missingExists)
	}

	xpathNodes, err := target.XPath(ctx, runtime.NewString("//article"))
	if err != nil {
		t.Fatalf("xpath node query failed: %v", err)
	}

	xpathNodeList, ok := xpathNodes.(runtime.List)
	if !ok {
		t.Fatalf("expected xpath node query to return a list, got %T", xpathNodes)
	}

	assertElementIDs(t, ctx, xpathNodeList, []string{"hero", "secondary"})

	xpathAttrs, err := target.XPath(ctx, runtime.NewString("//a/@title"))
	if err != nil {
		t.Fatalf("xpath attribute query failed: %v", err)
	}

	attrList, ok := xpathAttrs.(runtime.List)
	if !ok {
		t.Fatalf("expected xpath attribute query to return a list, got %T", xpathAttrs)
	}

	attrValue, err := attrList.At(ctx, runtime.NewInt(0))
	if err != nil {
		t.Fatalf("failed to read first xpath attribute result: %v", err)
	}

	if runtime.CompareValues(attrValue, runtime.NewString("Continue")) != 0 {
		t.Fatalf("expected xpath attribute value Continue, got %v", attrValue)
	}

	xpathCount, err := target.XPath(ctx, runtime.NewString("count(//article)"))
	if err != nil {
		t.Fatalf("xpath count query failed: %v", err)
	}

	if runtime.CompareValues(xpathCount, runtime.NewFloat(2)) != 0 {
		t.Fatalf("expected xpath count 2, got %v", xpathCount)
	}
}

func assertSharedReadSemantics(t *testing.T, ctx context.Context, target drivers.QueryTarget) {
	t.Helper()

	heroValue, err := target.QuerySelector(ctx, drivers.NewCSSSelector("#hero"))
	if err != nil {
		t.Fatalf("failed to resolve hero element: %v", err)
	}

	hero := mustElementFromValue(t, heroValue)

	contentTarget, err := drivers.ToContentTarget(hero)
	if err != nil {
		t.Fatalf("expected content capability: %v", err)
	}

	innerText, err := contentTarget.GetInnerText(ctx)
	if err != nil {
		t.Fatalf("get inner text failed: %v", err)
	}

	if runtime.CompareValues(innerText, runtime.NewString("FerretBrowser automationNext")) != 0 {
		t.Fatalf("unexpected inner text: %v", innerText)
	}

	innerHTML, err := contentTarget.GetInnerHTML(ctx)
	if err != nil {
		t.Fatalf("get inner html failed: %v", err)
	}

	if runtime.CompareValues(innerHTML, runtime.NewString(`<h1>Ferret</h1><p class="lead">Browser automation</p><a class="action" href="/next" title="Continue">Next</a><input id="search" value="find me"/>`)) != 0 {
		t.Fatalf("unexpected inner html: %v", innerHTML)
	}

	attributeTarget, err := drivers.ToAttributeTarget(hero)
	if err != nil {
		t.Fatalf("expected attribute capability: %v", err)
	}

	attrs, err := attributeTarget.GetAttributes(ctx)
	if err != nil {
		t.Fatalf("get attributes failed: %v", err)
	}

	assertMapStringValue(t, ctx, attrs, "data-role", "hero")
	assertMapStringValue(t, ctx, attrs, "style", "color:red;display:block")

	attrValue, err := attributeTarget.GetAttribute(ctx, runtime.NewString("data-role"))
	if err != nil {
		t.Fatalf("get attribute failed: %v", err)
	}

	if runtime.CompareValues(attrValue, runtime.NewString("hero")) != 0 {
		t.Fatalf("unexpected data-role attribute value: %v", attrValue)
	}

	styleTarget, err := drivers.ToStyleTarget(hero)
	if err != nil {
		t.Fatalf("expected style capability: %v", err)
	}

	styles, err := styleTarget.GetStyles(ctx)
	if err != nil {
		t.Fatalf("get styles failed: %v", err)
	}

	assertMapStringValue(t, ctx, styles, "color", "red")
	assertMapStringValue(t, ctx, styles, "display", "block")

	styleValue, err := styleTarget.GetStyle(ctx, runtime.NewString("color"))
	if err != nil {
		t.Fatalf("get style failed: %v", err)
	}

	if runtime.CompareValues(styleValue, runtime.NewString("red")) != 0 {
		t.Fatalf("unexpected style value: %v", styleValue)
	}

	searchValue, err := target.QuerySelector(ctx, drivers.NewCSSSelector("#search"))
	if err != nil {
		t.Fatalf("failed to resolve search element: %v", err)
	}

	search := mustElementFromValue(t, searchValue)
	valueTarget, err := drivers.ToValueTarget(search)
	if err != nil {
		t.Fatalf("expected value capability: %v", err)
	}

	value, err := valueTarget.GetValue(ctx)
	if err != nil {
		t.Fatalf("get value failed: %v", err)
	}

	if runtime.CompareValues(value, runtime.NewString("find me")) != 0 {
		t.Fatalf("unexpected input value: %v", value)
	}
}

func assertElementIDs(t *testing.T, ctx context.Context, list runtime.List, expected []string) {
	t.Helper()

	length, err := list.Length(ctx)
	if err != nil {
		t.Fatalf("failed to read list length: %v", err)
	}

	if runtime.CompareValues(length, runtime.NewInt(len(expected))) != 0 {
		t.Fatalf("expected %d elements, got %v", len(expected), length)
	}

	for idx, id := range expected {
		value, err := list.At(ctx, runtime.NewInt(idx))
		if err != nil {
			t.Fatalf("failed to read list item %d: %v", idx, err)
		}

		element := mustElementFromValue(t, value)
		assertElementID(t, ctx, element, id)
	}
}

func assertElementID(t *testing.T, ctx context.Context, element drivers.HTMLElement, expected string) {
	t.Helper()

	attrs, err := drivers.ToAttributeTarget(element)
	if err != nil {
		t.Fatalf("expected attribute capability: %v", err)
	}

	id, err := attrs.GetAttribute(ctx, runtime.NewString("id"))
	if err != nil {
		t.Fatalf("failed to read id attribute: %v", err)
	}

	if runtime.CompareValues(id, runtime.NewString(expected)) != 0 {
		t.Fatalf("expected id %q, got %v", expected, id)
	}
}

func assertMapStringValue(t *testing.T, ctx context.Context, value runtime.Map, key, expected string) {
	t.Helper()

	actual, err := value.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("failed to read map key %q: %v", key, err)
	}

	if runtime.CompareValues(actual, runtime.NewString(expected)) != 0 {
		t.Fatalf("expected %q=%q, got %v", key, expected, actual)
	}
}

func mustElementFromValue(t *testing.T, value runtime.Value) drivers.HTMLElement {
	t.Helper()

	element, err := drivers.ToElement(value)
	if err != nil {
		t.Fatalf("expected element value, got %T: %v", value, err)
	}

	return element
}
