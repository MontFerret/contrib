package drivers_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp"
	cdpdom "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

var (
	_ drivers.HTMLPage           = (*memory.HTMLPage)(nil)
	_ drivers.PageCookieReader   = (*memory.HTMLPage)(nil)
	_ drivers.PageResponseTarget = (*memory.HTMLPage)(nil)

	_ drivers.HTMLDocument           = (*memory.HTMLDocument)(nil)
	_ drivers.QueryTarget            = (*memory.HTMLDocument)(nil)
	_ drivers.DocumentMetadataTarget = (*memory.HTMLDocument)(nil)

	_ drivers.HTMLElement     = (*memory.HTMLElement)(nil)
	_ drivers.QueryTarget     = (*memory.HTMLElement)(nil)
	_ drivers.ContentTarget   = (*memory.HTMLElement)(nil)
	_ drivers.AttributeTarget = (*memory.HTMLElement)(nil)
	_ drivers.StyleTarget     = (*memory.HTMLElement)(nil)
	_ drivers.ValueTarget     = (*memory.HTMLElement)(nil)
	_ drivers.RelationTarget  = (*memory.HTMLElement)(nil)

	_ drivers.HTMLPage             = (*cdp.HTMLPage)(nil)
	_ drivers.PageCookieTarget     = (*cdp.HTMLPage)(nil)
	_ drivers.PageResponseTarget   = (*cdp.HTMLPage)(nil)
	_ drivers.PageSnapshotTarget   = (*cdp.HTMLPage)(nil)
	_ drivers.PageNavigationTarget = (*cdp.HTMLPage)(nil)

	_ drivers.HTMLDocument           = (*cdpdom.HTMLDocument)(nil)
	_ drivers.QueryTarget            = (*cdpdom.HTMLDocument)(nil)
	_ drivers.DocumentMetadataTarget = (*cdpdom.HTMLDocument)(nil)
	_ drivers.DocumentViewportTarget = (*cdpdom.HTMLDocument)(nil)

	_ drivers.HTMLElement       = (*cdpdom.HTMLElement)(nil)
	_ drivers.QueryTarget       = (*cdpdom.HTMLElement)(nil)
	_ drivers.ContentTarget     = (*cdpdom.HTMLElement)(nil)
	_ drivers.AttributeTarget   = (*cdpdom.HTMLElement)(nil)
	_ drivers.StyleTarget       = (*cdpdom.HTMLElement)(nil)
	_ drivers.ValueTarget       = (*cdpdom.HTMLElement)(nil)
	_ drivers.RelationTarget    = (*cdpdom.HTMLElement)(nil)
	_ drivers.InteractionTarget = (*cdpdom.HTMLElement)(nil)
	_ drivers.WaitTarget        = (*cdpdom.HTMLElement)(nil)
)

func TestRoleResolversAreExact(t *testing.T) {
	t.Parallel()

	page := newMemoryPage(t, `<html><body><div id="root"></div></body></html>`)
	doc := page.GetMainFrame()

	if _, err := drivers.ToDocument(page); err == nil {
		t.Fatal("expected page -> document cast to fail")
	}

	if _, err := drivers.ToElement(page); err == nil {
		t.Fatal("expected page -> element cast to fail")
	}

	if _, err := drivers.ToElement(doc); err == nil {
		t.Fatal("expected document -> element cast to fail")
	}

	if _, err := drivers.ToDocument(doc); err != nil {
		t.Fatalf("expected exact document cast to succeed: %v", err)
	}
}

func TestElementCapabilityResolversAreExact(t *testing.T) {
	t.Parallel()

	doc := newCapabilityDocument(t, `<html><body><button id="cta">go</button></body></html>`)
	page := newCapabilityPage(t, doc)
	element := doc.GetElement()

	if _, err := drivers.ToInteractionTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected page interaction capability to stay exact, got %v", err)
	}

	if _, err := drivers.ToWaitTarget(doc); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected document wait capability to stay exact, got %v", err)
	}

	if _, err := drivers.ToContentTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected page content capability to stay exact, got %v", err)
	}

	if _, err := drivers.ToAttributeTarget(doc); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected document attribute capability to stay exact, got %v", err)
	}

	if _, err := drivers.ToInteractionTarget(element); err != nil {
		t.Fatalf("expected element interaction capability: %v", err)
	}

	if _, err := drivers.ToWaitTarget(element); err != nil {
		t.Fatalf("expected element wait capability: %v", err)
	}

	if _, err := drivers.ToContentTarget(element); err != nil {
		t.Fatalf("expected element content capability: %v", err)
	}

	if _, err := drivers.ToAttributeTarget(element); err != nil {
		t.Fatalf("expected element attribute capability: %v", err)
	}
}

func TestDocumentCapabilityResolversStillCoerceFromPage(t *testing.T) {
	t.Parallel()

	doc := newCapabilityDocument(t, `<html><body><button id="cta">go</button></body></html>`)
	page := newCapabilityPage(t, doc)
	ctx := context.Background()

	viewport, err := drivers.ToDocumentViewportTarget(page)
	if err != nil {
		t.Fatalf("expected page viewport capability: %v", err)
	}

	if err := viewport.ScrollTop(ctx, drivers.ScrollOptions{}); err != nil {
		t.Fatalf("unexpected scroll error: %v", err)
	}

	if !doc.scrolledTop {
		t.Fatal("expected page viewport resolver to delegate to the main frame")
	}
}

func TestCapabilityResolversRejectUnsupportedBackends(t *testing.T) {
	t.Parallel()

	page := newMemoryPage(t, `<html><body><div></div></body></html>`)

	if _, err := drivers.ToInteractionTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected unsupported interaction capability error, got %v", err)
	}

	if _, err := drivers.ToWaitTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected unsupported wait capability error, got %v", err)
	}

	if _, err := drivers.ToContentTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected unsupported content capability error, got %v", err)
	}

	if _, err := drivers.ToDocumentViewportTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected unsupported viewport capability error, got %v", err)
	}

	if _, err := drivers.ToPageNavigationTarget(page); !errors.Is(err, runtime.ErrNotSupported) {
		t.Fatalf("expected unsupported navigation capability error, got %v", err)
	}

	if _, err := drivers.ToPageCookieReader(page); err != nil {
		t.Fatalf("expected cookie capability on memory page: %v", err)
	}
}

type capabilityPage struct {
	*memory.HTMLPage
	frame drivers.HTMLDocument
}

func (p *capabilityPage) GetMainFrame() drivers.HTMLDocument {
	return p.frame
}

type capabilityDocument struct {
	*memory.HTMLDocument
	element     *capabilityElement
	scrolledTop bool
}

func (doc *capabilityDocument) GetElement() drivers.HTMLElement {
	return doc.element
}

func (doc *capabilityDocument) Scroll(_ context.Context, _ drivers.ScrollOptions) error {
	return nil
}

func (doc *capabilityDocument) ScrollTop(_ context.Context, _ drivers.ScrollOptions) error {
	doc.scrolledTop = true
	return nil
}

func (doc *capabilityDocument) ScrollBottom(_ context.Context, _ drivers.ScrollOptions) error {
	return nil
}

func (doc *capabilityDocument) ScrollBySelector(_ context.Context, _ drivers.QuerySelector, _ drivers.ScrollOptions) error {
	return nil
}

func (doc *capabilityDocument) MoveMouseByXY(_ context.Context, _, _ runtime.Float) error {
	return nil
}

type capabilityElement struct {
	*memory.HTMLElement
	waitSelector string
	clickCount   runtime.Int
}

func (el *capabilityElement) Click(_ context.Context, count runtime.Int) error {
	el.clickCount = count
	return nil
}

func (el *capabilityElement) ClickBySelector(_ context.Context, _ drivers.QuerySelector, count runtime.Int) error {
	el.clickCount = count
	return nil
}

func (el *capabilityElement) ClickBySelectorAll(_ context.Context, _ drivers.QuerySelector, count runtime.Int) error {
	el.clickCount = count
	return nil
}

func (el *capabilityElement) Clear(_ context.Context) error {
	return nil
}

func (el *capabilityElement) ClearBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *capabilityElement) Input(_ context.Context, _ runtime.Value, _ runtime.Int) error {
	return nil
}

func (el *capabilityElement) InputBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.Value, _ runtime.Int) error {
	return nil
}

func (el *capabilityElement) Press(_ context.Context, _ []runtime.String, _ runtime.Int) error {
	return nil
}

func (el *capabilityElement) PressBySelector(_ context.Context, _ drivers.QuerySelector, _ []runtime.String, _ runtime.Int) error {
	return nil
}

func (el *capabilityElement) Select(_ context.Context, value runtime.List) (runtime.List, error) {
	return value, nil
}

func (el *capabilityElement) SelectBySelector(_ context.Context, _ drivers.QuerySelector, value runtime.List) (runtime.List, error) {
	return value, nil
}

func (el *capabilityElement) ScrollIntoView(_ context.Context, _ drivers.ScrollOptions) error {
	return nil
}

func (el *capabilityElement) Focus(_ context.Context) error {
	return nil
}

func (el *capabilityElement) FocusBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *capabilityElement) Blur(_ context.Context) error {
	return nil
}

func (el *capabilityElement) BlurBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *capabilityElement) Hover(_ context.Context) error {
	return nil
}

func (el *capabilityElement) HoverBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *capabilityElement) WaitForElement(_ context.Context, selector drivers.QuerySelector, _ drivers.WaitEvent) error {
	el.waitSelector = selector.String()
	return nil
}

func (el *capabilityElement) WaitForElementAll(_ context.Context, _ drivers.QuerySelector, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForAttribute(_ context.Context, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForAttributeBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForAttributeBySelectorAll(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForStyle(_ context.Context, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForStyleBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForStyleBySelectorAll(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForClass(_ context.Context, _ runtime.String, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForClassBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ drivers.WaitEvent) error {
	return nil
}

func (el *capabilityElement) WaitForClassBySelectorAll(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ drivers.WaitEvent) error {
	return nil
}

func newCapabilityPage(t *testing.T, doc *capabilityDocument) *capabilityPage {
	t.Helper()

	page := newMemoryPage(t, `<html><body><div></div></body></html>`)

	return &capabilityPage{
		HTMLPage: page,
		frame:    doc,
	}
}

func newCapabilityDocument(t *testing.T, markup string) *capabilityDocument {
	t.Helper()

	base := newMemoryDocument(t, markup)
	element, ok := base.GetElement().(*memory.HTMLElement)
	if !ok {
		t.Fatalf("expected memory element, got %T", base.GetElement())
	}

	return &capabilityDocument{
		HTMLDocument: base,
		element: &capabilityElement{
			HTMLElement: element,
		},
	}
}

func newMemoryPage(t *testing.T, markup string) *memory.HTMLPage {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
	if err != nil {
		t.Fatalf("failed to parse document: %v", err)
	}

	page, err := memory.NewHTMLPage(
		doc,
		"https://example.com",
		drivers.HTTPResponse{Headers: drivers.NewHTTPHeaders()},
		drivers.NewHTTPCookies(),
	)
	if err != nil {
		t.Fatalf("failed to create memory page: %v", err)
	}

	return page
}

func newMemoryDocument(t *testing.T, markup string) *memory.HTMLDocument {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
	if err != nil {
		t.Fatalf("failed to parse document: %v", err)
	}

	out, err := memory.NewRootHTMLDocument(doc, "https://example.com")
	if err != nil {
		t.Fatalf("failed to create memory document: %v", err)
	}

	return out
}
