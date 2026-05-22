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
	cases := []struct {
		typ      reflect.Type
		name     string
		writable bool
	}{
		{name: "memory page", typ: reflect.TypeOf((*memory.HTMLPage)(nil))},
		{name: "memory document", typ: reflect.TypeOf((*memory.HTMLDocument)(nil))},
		{name: "memory element", typ: reflect.TypeOf((*memory.HTMLElement)(nil))},
		{name: "cdp page", typ: reflect.TypeOf((*cdp.HTMLPage)(nil))},
		{name: "cdp document", typ: reflect.TypeOf((*cdpdom.HTMLDocument)(nil))},
		{name: "cdp element", typ: reflect.TypeOf((*cdpdom.HTMLElement)(nil)), writable: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.typ.Implements(keyWritable); got != tc.writable {
				t.Fatalf("%s KeyWritable support mismatch: got %t, want %t", tc.name, got, tc.writable)
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
