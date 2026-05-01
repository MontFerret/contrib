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

func TestHTMLValuesDoNotImplementKeyWritable(t *testing.T) {
	t.Parallel()

	keyWritable := reflect.TypeOf((*runtime.KeyWritable)(nil)).Elem()
	cases := []struct {
		typ  reflect.Type
		name string
	}{
		{name: "memory page", typ: reflect.TypeOf((*memory.HTMLPage)(nil))},
		{name: "memory document", typ: reflect.TypeOf((*memory.HTMLDocument)(nil))},
		{name: "memory element", typ: reflect.TypeOf((*memory.HTMLElement)(nil))},
		{name: "cdp page", typ: reflect.TypeOf((*cdp.HTMLPage)(nil))},
		{name: "cdp document", typ: reflect.TypeOf((*cdpdom.HTMLDocument)(nil))},
		{name: "cdp element", typ: reflect.TypeOf((*cdpdom.HTMLElement)(nil))},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.typ.Implements(keyWritable) {
				t.Fatalf("%s unexpectedly implements runtime.KeyWritable", tc.name)
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

	if err := hero.SetAttribute(ctx, runtime.NewString("data-role"), runtime.NewString("hero")); err != nil {
		t.Fatalf("set attribute: %v", err)
	}

	attrValue, err := hero.GetAttribute(ctx, runtime.NewString("data-role"))
	if err != nil {
		t.Fatalf("get attribute: %v", err)
	}

	if runtime.CompareValues(attrValue, runtime.NewString("hero")) != 0 {
		t.Fatalf("unexpected attribute value: %v", attrValue)
	}

	if err := hero.SetStyle(ctx, runtime.NewString("color"), runtime.NewString("red")); err != nil {
		t.Fatalf("set style: %v", err)
	}

	styleValue, err := hero.GetStyle(ctx, runtime.NewString("color"))
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

	if err := search.SetValue(ctx, runtime.NewString("updated")); err != nil {
		t.Fatalf("set value: %v", err)
	}

	value, err := search.GetValue(ctx)
	if err != nil {
		t.Fatalf("get value: %v", err)
	}

	if runtime.CompareValues(value, runtime.NewString("updated")) != 0 {
		t.Fatalf("unexpected input value: %v", value)
	}
}
