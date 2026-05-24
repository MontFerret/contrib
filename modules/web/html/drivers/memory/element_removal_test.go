package memory_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestHTMLElementRemoveAtRemovesChildAndRefreshesChildren(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	for _, tc := range []struct {
		name          string
		removedText   string
		nextIndexAttr string
		idx           runtime.Int
		nextIdx       runtime.Int
	}{
		{name: "first child", idx: 0, removedText: "one", nextIdx: 0, nextIndexAttr: "1"},
		{name: "middle child", idx: 1, removedText: "two", nextIdx: 1, nextIndexAttr: "2"},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			el := newRemovalElement(t)
			removable := el.(runtime.IndexRemovable)

			removed, err := removable.RemoveAt(ctx, tc.idx)
			if err != nil {
				t.Fatalf("remove child: %v", err)
			}

			if got := elementInnerText(t, ctx, removed); got != tc.removedText {
				t.Fatalf("unexpected removed child text: got %q, want %q", got, tc.removedText)
			}

			length, err := el.Length(ctx)
			if err != nil {
				t.Fatalf("length: %v", err)
			}

			if length != 2 {
				t.Fatalf("unexpected length after removal: got %d, want 2", length)
			}

			next, err := el.GetChildNode(ctx, tc.nextIdx)
			if err != nil {
				t.Fatalf("next child: %v", err)
			}

			if got := elementAttribute(t, ctx, next, "data-index"); got != tc.nextIndexAttr {
				t.Fatalf("unexpected shifted child data-index: got %q, want %q", got, tc.nextIndexAttr)
			}
		})
	}
}

func TestHTMLElementRemoveAtMissingIndexReturnsNone(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	el := newRemovalElement(t)
	removable := el.(runtime.IndexRemovable)

	removed, err := removable.RemoveAt(ctx, 10)
	if err != nil {
		t.Fatalf("remove missing child: %v", err)
	}

	if removed != runtime.None {
		t.Fatalf("expected missing child removal to return none, got %v", removed)
	}

	length, err := el.Length(ctx)
	if err != nil {
		t.Fatalf("length: %v", err)
	}

	if length != 3 {
		t.Fatalf("unexpected length after missing removal: got %d, want 3", length)
	}
}

func TestHTMLElementRemoveKeyAcceptsOnlyIntegerKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	el := newRemovalElement(t)
	removable := el.(runtime.KeyRemovable)

	if err := removable.RemoveKey(ctx, runtime.NewString("attributes")); !errors.Is(err, runtime.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument error, got %v", err)
	}

	if err := removable.RemoveKey(ctx, runtime.NewInt(0)); err != nil {
		t.Fatalf("remove by integer key: %v", err)
	}

	length, err := el.Length(ctx)
	if err != nil {
		t.Fatalf("length: %v", err)
	}

	if length != 2 {
		t.Fatalf("unexpected length after integer-key removal: got %d, want 2", length)
	}
}

func newRemovalElement(t *testing.T) drivers.HTMLElement {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(`
		<html>
			<body>
				<ul id="list">
					<li data-index="0">one</li>
					<li data-index="1">two</li>
					<li data-index="2">three</li>
				</ul>
			</body>
		</html>
	`))
	if err != nil {
		t.Fatalf("new document: %v", err)
	}

	el, err := memory.NewHTMLElement(doc, doc.Find("#list"))
	if err != nil {
		t.Fatalf("new element: %v", err)
	}

	return el
}

func elementInnerText(t *testing.T, ctx context.Context, value runtime.Value) string {
	t.Helper()

	el, err := drivers.ToElement(value)
	if err != nil {
		t.Fatalf("to element: %v", err)
	}

	target, err := drivers.ToContentTarget(el)
	if err != nil {
		t.Fatalf("content target: %v", err)
	}

	text, err := target.GetInnerText(ctx)
	if err != nil {
		t.Fatalf("inner text: %v", err)
	}

	return text.String()
}

func elementAttribute(t *testing.T, ctx context.Context, value runtime.Value, name string) string {
	t.Helper()

	el, err := drivers.ToElement(value)
	if err != nil {
		t.Fatalf("to element: %v", err)
	}

	target, err := drivers.ToAttributeTarget(el)
	if err != nil {
		t.Fatalf("attribute target: %v", err)
	}

	attr, err := target.GetAttribute(ctx, runtime.NewString(name))
	if err != nil {
		t.Fatalf("attribute %s: %v", name, err)
	}

	return attr.String()
}
