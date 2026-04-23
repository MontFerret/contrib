package drivers_test

import (
	"context"
	"testing"

	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const cssxContractMarkup = `<html><body><section class="card" data-role="hero"><h1>Hero</h1><a href="/hero">Read</a></section><section class="card" data-role="related"><h2>Other</h2><span class="price">$1,234.50</span></section></body></html>`

func TestCSSXContractAcrossBackends(t *testing.T) {
	t.Parallel()

	doc := newMemoryDocument(t, cssxContractMarkup)
	root := doc.GetElement().(*memory.HTMLElement)
	ctx := context.Background()

	cases := []struct {
		assert func(*testing.T, runtime.List)
		name   string
		exp    string
	}{
		{
			name: "first card",
			exp:  `:first(.card)`,
			assert: func(t *testing.T, list runtime.List) {
				length, err := list.Length(ctx)
				if err != nil {
					t.Fatalf("list length: %v", err)
				}
				if length != 1 {
					t.Fatalf("expected one element, got %d", length)
				}

				first, err := list.At(ctx, runtime.NewInt(0))
				if err != nil {
					t.Fatalf("read first item: %v", err)
				}
				element := mustElementFromValue(t, first)
				attr, err := element.GetAttribute(ctx, runtime.NewString("data-role"))
				if err != nil {
					t.Fatalf("read data-role: %v", err)
				}
				if runtime.CompareValues(attr, runtime.NewString("hero")) != 0 {
					t.Fatalf("unexpected data-role: %v", attr)
				}
			},
		},
		{
			name: "card attrs",
			exp:  `:attrs("data-role", .card)`,
			assert: func(t *testing.T, list runtime.List) {
				assertListValues(t, ctx, list, []runtime.Value{runtime.NewString("hero"), runtime.NewString("related")})
			},
		},
		{
			name: "count cards",
			exp:  `:count(.card)`,
			assert: func(t *testing.T, list runtime.List) {
				assertListValues(t, ctx, list, []runtime.Value{runtime.NewInt(2)})
			},
		},
		{
			name: "price to number",
			exp:  `:toNumber(:text(:first(.price)))`,
			assert: func(t *testing.T, list runtime.List) {
				assertListValues(t, ctx, list, []runtime.Value{runtime.NewFloat(1234.5)})
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			list, err := memory.EvalCSSX(ctx, root, runtime.NewString(tc.exp))
			if err != nil {
				t.Fatalf("memory eval failed: %v", err)
			}

			tc.assert(t, list)

			fn, err := templates.CSSX(cdpruntime.RemoteObjectID("obj"), runtime.NewString(tc.exp))
			if err != nil {
				t.Fatalf("cdp compile failed: %v", err)
			}

			if fn == nil {
				t.Fatal("expected cdp function")
			}
		})
	}
}

func assertListValues(t *testing.T, ctx context.Context, list runtime.List, expected []runtime.Value) {
	t.Helper()

	length, err := list.Length(ctx)
	if err != nil {
		t.Fatalf("list length: %v", err)
	}

	if int(length) != len(expected) {
		t.Fatalf("expected %d values, got %d", len(expected), length)
	}

	for idx, want := range expected {
		got, err := list.At(ctx, runtime.NewInt(idx))
		if err != nil {
			t.Fatalf("list[%d]: %v", idx, err)
		}

		if runtime.CompareValues(got, want) != 0 {
			t.Fatalf("list[%d]: expected %v, got %v", idx, want, got)
		}
	}
}
