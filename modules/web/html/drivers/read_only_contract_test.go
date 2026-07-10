package drivers_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
	cdpdom "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestOnlyCDPElementImplementsKeyWritable(t *testing.T) {
	t.Parallel()

	keyWritable := reflect.TypeOf((*runtime.KeyWritable)(nil)).Elem()
	indexRemovable := reflect.TypeOf((*runtime.IndexRemovable)(nil)).Elem()
	keyRemovable := reflect.TypeOf((*runtime.KeyRemovable)(nil)).Elem()
	cases := []struct {
		typ            reflect.Type
		name           string
		writable       bool
		indexRemovable bool
		keyRemovable   bool
	}{
		{name: "memory page", typ: reflect.TypeOf((*memory.HTMLPage)(nil))},
		{name: "memory document", typ: reflect.TypeOf((*memory.HTMLDocument)(nil))},
		{name: "memory element", typ: reflect.TypeOf((*memory.HTMLElement)(nil)), indexRemovable: true, keyRemovable: true},
		{name: "cdp page", typ: reflect.TypeOf((*cdp.HTMLPage)(nil))},
		{name: "cdp document", typ: reflect.TypeOf((*cdpdom.HTMLDocument)(nil))},
		{name: "cdp element", typ: reflect.TypeOf((*cdpdom.HTMLElement)(nil)), writable: true, indexRemovable: true, keyRemovable: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.typ.Implements(keyWritable); got != tc.writable {
				t.Fatalf("%s KeyWritable support mismatch: got %t, want %t", tc.name, got, tc.writable)
			}

			if got := tc.typ.Implements(indexRemovable); got != tc.indexRemovable {
				t.Fatalf("%s IndexRemovable support mismatch: got %t, want %t", tc.name, got, tc.indexRemovable)
			}

			if got := tc.typ.Implements(keyRemovable); got != tc.keyRemovable {
				t.Fatalf("%s KeyRemovable support mismatch: got %t, want %t", tc.name, got, tc.keyRemovable)
			}
		})
	}
}

func TestDocumentAndPageDotAccessDoNotExposeElementOnlyViews(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	page := newMemoryPage(t, `<html><body><article id="hero" class="card" style="display:block" data-role="hero">Text</article></body></html>`)
	doc := page.GetMainFrame()

	for _, tc := range []struct {
		target runtime.KeyReadable
		key    string
		name   string
	}{
		{name: "document attributes", target: doc, key: "attributes"},
		{name: "document style", target: doc, key: "style"},
		{name: "document value", target: doc, key: "value"},
		{name: "page attributes", target: page, key: "attributes"},
		{name: "page style", target: page, key: "style"},
		{name: "page value", target: page, key: "value"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			value, err := tc.target.Get(ctx, runtime.NewString(tc.key))
			if err != nil {
				t.Fatalf("read %s: %v", tc.key, err)
			}

			if value != runtime.None {
				t.Fatalf("expected %s to stay element-only, got %T %v", tc.key, value, value)
			}
		})
	}

	for _, tc := range []struct {
		target runtime.KeyReadable
		name   string
	}{
		{name: "document", target: doc},
		{name: "page", target: page},
	} {
		tc := tc
		t.Run(tc.name+" inner content", func(t *testing.T) {
			t.Parallel()

			value, err := tc.target.Get(ctx, runtime.NewString("innerText"))
			if err != nil {
				t.Fatalf("read innerText: %v", err)
			}

			if runtime.CompareValues(value, runtime.NewString("Text")) != 0 {
				t.Fatalf("expected innerText to remain readable, got %v", value)
			}
		})
	}
}

func TestMemoryElementRejectsNonStandardTextAliases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	page := newMemoryPage(t, `<html><body><article id="hero">Text</article></body></html>`)
	doc := page.GetMainFrame()

	heroValue, err := doc.QuerySelector(ctx, drivers.NewCSSSelector("#hero"))
	if err != nil {
		t.Fatalf("resolve hero: %v", err)
	}

	hero := mustElementFromValue(t, heroValue)

	for _, name := range []string{"text", "html"} {
		name := name

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, err := hero.Get(ctx, runtime.NewString(name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}

			if value != runtime.None {
				t.Fatalf("expected %s read to return none, got %T %v", name, value, value)
			}
		})
	}
}

func TestMemoryElementDOMPropertyFallback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	doc := newMemoryDocument(t, `<html><body><select id="choices" class="picker" multiple disabled><option value="1" selected>One</option><option value="2">Two</option><optgroup><option value="3" selected>Three</option></optgroup></select></body></html>`)

	choicesValue, err := doc.QuerySelector(ctx, drivers.NewCSSSelector("#choices"))
	if err != nil {
		t.Fatalf("resolve choices: %v", err)
	}

	choices := mustElementFromValue(t, choicesValue)
	assertElementRead(t, ctx, choices, "id", runtime.NewString("choices"))
	assertElementRead(t, ctx, choices, "className", runtime.NewString("picker"))
	assertElementRead(t, ctx, choices, "disabled", runtime.True)
	assertElementRead(t, ctx, choices, "nodeName", runtime.NewString("select"))

	firstChild, err := choices.Get(ctx, runtime.NewInt(0))
	if err != nil {
		t.Fatalf("read first child: %v", err)
	}

	firstOption := mustElementFromValue(t, firstChild)
	assertElementRead(t, ctx, firstOption, "value", runtime.NewString("1"))

	selectedValue, err := choices.Get(ctx, runtime.NewString("selectedOptions"))
	if err != nil {
		t.Fatalf("read selectedOptions: %v", err)
	}

	selected, ok := selectedValue.(runtime.List)
	if !ok {
		t.Fatalf("expected selectedOptions list, got %T", selectedValue)
	}

	assertElementValues(t, ctx, selected, []string{"1", "3"})
	assertElementRead(t, ctx, firstOption, "selected", runtime.True)
	assertElementRead(t, ctx, choices, "unknownProperty", runtime.None)
}

func TestExplicitElementMutationCapabilitiesRemain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	page := newMemoryPage(t, `<html><body><article id="hero"></article><input id="search" value="find me"/></body></html>`)
	doc := page.GetMainFrame()

	heroValue, err := doc.QuerySelector(ctx, drivers.NewCSSSelector("#hero"))
	if err != nil {
		t.Fatalf("resolve hero: %v", err)
	}

	hero := mustElementFromValue(t, heroValue)
	heroAttrs, err := drivers.ToAttributeTarget(hero)
	if err != nil {
		t.Fatalf("hero attribute target: %v", err)
	}

	if err := heroAttrs.SetAttribute(ctx, runtime.NewString("data-role"), runtime.NewString("hero")); err != nil {
		t.Fatalf("set attribute: %v", err)
	}

	attrValue, err := heroAttrs.GetAttribute(ctx, runtime.NewString("data-role"))
	if err != nil {
		t.Fatalf("get attribute: %v", err)
	}

	if runtime.CompareValues(attrValue, runtime.NewString("hero")) != 0 {
		t.Fatalf("unexpected attribute value: %v", attrValue)
	}

	heroStyles, err := drivers.ToStyleTarget(hero)
	if err != nil {
		t.Fatalf("hero style target: %v", err)
	}

	if err := heroStyles.SetStyle(ctx, runtime.NewString("color"), runtime.NewString("red")); err != nil {
		t.Fatalf("set style: %v", err)
	}

	styleValue, err := heroStyles.GetStyle(ctx, runtime.NewString("color"))
	if err != nil {
		t.Fatalf("get style: %v", err)
	}

	if runtime.CompareValues(styleValue, runtime.NewString("red")) != 0 {
		t.Fatalf("unexpected style value: %v", styleValue)
	}

	searchValue, err := doc.QuerySelector(ctx, drivers.NewCSSSelector("#search"))
	if err != nil {
		t.Fatalf("resolve search: %v", err)
	}

	search := mustElementFromValue(t, searchValue)
	searchValueTarget, err := drivers.ToValueTarget(search)
	if err != nil {
		t.Fatalf("search value target: %v", err)
	}

	if err := searchValueTarget.SetValue(ctx, runtime.NewString("updated")); err != nil {
		t.Fatalf("set value: %v", err)
	}

	value, err := searchValueTarget.GetValue(ctx)
	if err != nil {
		t.Fatalf("get value: %v", err)
	}

	if runtime.CompareValues(value, runtime.NewString("updated")) != 0 {
		t.Fatalf("unexpected input value: %v", value)
	}
}

func assertElementRead(t *testing.T, ctx context.Context, element drivers.HTMLElement, key string, expected runtime.Value) {
	t.Helper()

	value, err := element.Get(ctx, runtime.NewString(key))
	if err != nil {
		t.Fatalf("read %s: %v", key, err)
	}

	if runtime.CompareValues(value, expected) != 0 {
		t.Fatalf("expected %s to be %v, got %v", key, expected, value)
	}
}

func assertElementValues(t *testing.T, ctx context.Context, list runtime.List, expected []string) {
	t.Helper()

	length, err := list.Length(ctx)
	if err != nil {
		t.Fatalf("read selectedOptions length: %v", err)
	}

	if runtime.CompareValues(length, runtime.NewInt(len(expected))) != 0 {
		t.Fatalf("expected selectedOptions length %d, got %v", len(expected), length)
	}

	for idx, expectedValue := range expected {
		value, err := list.At(ctx, runtime.NewInt(idx))
		if err != nil {
			t.Fatalf("read selectedOptions[%d]: %v", idx, err)
		}

		option := mustElementFromValue(t, value)
		assertElementRead(t, ctx, option, "value", runtime.NewString(expectedValue))
	}
}
