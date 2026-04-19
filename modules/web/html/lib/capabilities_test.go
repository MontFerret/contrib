package lib

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/memory"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestGetInnerTextUsesContentCapabilityFromPage(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><div id="message">hello</div></body></html>`)

	value, err := GetInnerText(context.Background(), page, runtime.NewString("#message"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := value.String(); got != "hello" {
		t.Fatalf("expected inner text to resolve through page content capability, got %q", got)
	}
}

func TestClickUsesInteractionCapabilityFromPage(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><button id="cta">go</button></body></html>`)

	value, err := Click(context.Background(), page, runtime.NewString("#cta"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected click to report success, got %v", value)
	}

	if got := page.frame.element.clickedSelector; got != "#cta" {
		t.Fatalf("expected click selector to be delegated to the root element, got %q", got)
	}
}

func TestWaitElementUsesWaitCapabilityFromPage(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><button id="cta">go</button></body></html>`)

	value, err := WaitElement(context.Background(), page, runtime.NewString("#cta"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected wait to report success, got %v", value)
	}

	if got := page.frame.element.waitSelector; got != "#cta" {
		t.Fatalf("expected wait selector to be delegated to the root element, got %q", got)
	}
}

func TestScrollIntoUsesRootInteractionCapabilityFromDocument(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><div id="layout">viewport</div></body></html>`)

	value, err := ScrollInto(context.Background(), doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected scroll into view to report success, got %v", value)
	}

	if !doc.element.scrolledIntoView {
		t.Fatal("expected scroll into view to be delegated to the document root element")
	}
}

func TestScrollTopUsesViewportCapabilityFromPage(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><div>viewport</div></body></html>`)

	value, err := ScrollTop(context.Background(), page)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected scroll to report success, got %v", value)
	}

	if !page.frame.scrolledTop {
		t.Fatal("expected scroll to be delegated to the main frame viewport capability")
	}
}

func TestNavigateUsesPageNavigationCapability(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><div>nav</div></body></html>`)

	value, err := Navigate(context.Background(), page, runtime.NewString("https://next.example"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != runtime.True {
		t.Fatalf("expected navigate to report success, got %v", value)
	}

	if got := page.navigatedTo; got != "https://next.example" {
		t.Fatalf("expected navigation target to be recorded, got %q", got)
	}
}

func TestPDFUsesPageSnapshotCapability(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><div>pdf</div></body></html>`)

	value, err := PDF(context.Background(), page)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value == runtime.None {
		t.Fatal("expected PDF to return binary output")
	}

	if !page.printedPDF {
		t.Fatal("expected PDF helper to use page snapshot capability")
	}
}

func TestCookieGetUsesPageCookieCapability(t *testing.T) {
	t.Parallel()

	page := newMemoryPage(t, `<html><body><div>cookies</div></body></html>`, drivers.NewHTTPCookiesWith(map[string]drivers.HTTPCookie{
		"session": {
			Name:  "session",
			Value: "abc123",
		},
	}))

	value, err := CookieGet(context.Background(), page, runtime.NewString("session"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cookie, ok := runtime.UnwrapAs[drivers.HTTPCookie](value)
	if !ok {
		t.Fatalf("expected cookie value, got %T", value)
	}

	if cookie.Value != "abc123" {
		t.Fatalf("expected cookie value to round-trip, got %q", cookie.Value)
	}
}

func TestAttributeSetRemainsElementOnly(t *testing.T) {
	t.Parallel()

	page := newTestPage(t, `<html><body><div id="message"></div></body></html>`)

	if _, err := AttributeSet(context.Background(), page, runtime.NewString("data-test"), runtime.NewString("true")); err == nil {
		t.Fatal("expected page input to remain invalid for ATTR_SET")
	}
}

func TestPaginationRemainsPageOnly(t *testing.T) {
	t.Parallel()

	doc := newTestDocument(t, `<html><body><div id="message"></div></body></html>`)

	if _, err := Pagination(context.Background(), doc, runtime.NewString(".next")); err == nil {
		t.Fatal("expected document input to remain invalid for PAGINATION")
	}
}

type testPage struct {
	*memory.HTMLPage
	frame       *testDocument
	navigatedTo runtime.String
	printedPDF  bool
}

func (p *testPage) GetMainFrame() drivers.HTMLDocument {
	return p.frame
}

func (p *testPage) WaitForNavigation(_ context.Context, _ runtime.String) error {
	return nil
}

func (p *testPage) WaitForFrameNavigation(_ context.Context, _ drivers.HTMLDocument, _ runtime.String) error {
	return nil
}

func (p *testPage) Navigate(_ context.Context, url runtime.String) error {
	p.navigatedTo = url
	return nil
}

func (p *testPage) NavigateBack(_ context.Context, _ runtime.Int) (runtime.Boolean, error) {
	return runtime.True, nil
}

func (p *testPage) NavigateForward(_ context.Context, _ runtime.Int) (runtime.Boolean, error) {
	return runtime.True, nil
}

func (p *testPage) PrintToPDF(_ context.Context, _ drivers.PDFParams) (runtime.Binary, error) {
	p.printedPDF = true
	return runtime.NewBinary([]byte("pdf")), nil
}

func (p *testPage) CaptureScreenshot(_ context.Context, _ drivers.ScreenshotParams) (runtime.Binary, error) {
	return runtime.NewBinary([]byte("image")), nil
}

type testDocument struct {
	*memory.HTMLDocument
	element     *testElement
	scrolledTop bool
}

func (doc *testDocument) GetElement() drivers.HTMLElement {
	return doc.element
}

func (doc *testDocument) Scroll(_ context.Context, _ drivers.ScrollOptions) error {
	return nil
}

func (doc *testDocument) ScrollTop(_ context.Context, _ drivers.ScrollOptions) error {
	doc.scrolledTop = true
	return nil
}

func (doc *testDocument) ScrollBottom(_ context.Context, _ drivers.ScrollOptions) error {
	return nil
}

func (doc *testDocument) ScrollBySelector(_ context.Context, _ drivers.QuerySelector, _ drivers.ScrollOptions) error {
	return nil
}

func (doc *testDocument) MoveMouseByXY(_ context.Context, _, _ runtime.Float) error {
	return nil
}

type testElement struct {
	*memory.HTMLElement
	clickedSelector  string
	waitSelector     string
	scrolledIntoView bool
}

func (el *testElement) Click(_ context.Context, _ runtime.Int) error {
	return nil
}

func (el *testElement) ClickBySelector(_ context.Context, selector drivers.QuerySelector, _ runtime.Int) error {
	el.clickedSelector = selector.String()
	return nil
}

func (el *testElement) ClickBySelectorAll(_ context.Context, selector drivers.QuerySelector, _ runtime.Int) error {
	el.clickedSelector = selector.String()
	return nil
}

func (el *testElement) Clear(_ context.Context) error {
	return nil
}

func (el *testElement) ClearBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *testElement) Input(_ context.Context, _ runtime.Value, _ runtime.Int) error {
	return nil
}

func (el *testElement) InputBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.Value, _ runtime.Int) error {
	return nil
}

func (el *testElement) Press(_ context.Context, _ []runtime.String, _ runtime.Int) error {
	return nil
}

func (el *testElement) PressBySelector(_ context.Context, _ drivers.QuerySelector, _ []runtime.String, _ runtime.Int) error {
	return nil
}

func (el *testElement) Select(_ context.Context, value runtime.List) (runtime.List, error) {
	return value, nil
}

func (el *testElement) SelectBySelector(_ context.Context, _ drivers.QuerySelector, value runtime.List) (runtime.List, error) {
	return value, nil
}

func (el *testElement) ScrollIntoView(_ context.Context, _ drivers.ScrollOptions) error {
	el.scrolledIntoView = true
	return nil
}

func (el *testElement) Focus(_ context.Context) error {
	return nil
}

func (el *testElement) FocusBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *testElement) Blur(_ context.Context) error {
	return nil
}

func (el *testElement) BlurBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *testElement) Hover(_ context.Context) error {
	return nil
}

func (el *testElement) HoverBySelector(_ context.Context, _ drivers.QuerySelector) error {
	return nil
}

func (el *testElement) WaitForElement(_ context.Context, selector drivers.QuerySelector, _ drivers.WaitEvent) error {
	el.waitSelector = selector.String()
	return nil
}

func (el *testElement) WaitForElementAll(_ context.Context, _ drivers.QuerySelector, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForAttribute(_ context.Context, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForAttributeBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForAttributeBySelectorAll(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForStyle(_ context.Context, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForStyleBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForStyleBySelectorAll(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ runtime.Value, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForClass(_ context.Context, _ runtime.String, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForClassBySelector(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ drivers.WaitEvent) error {
	return nil
}

func (el *testElement) WaitForClassBySelectorAll(_ context.Context, _ drivers.QuerySelector, _ runtime.String, _ drivers.WaitEvent) error {
	return nil
}

func newTestPage(t *testing.T, markup string) *testPage {
	t.Helper()

	frame := newTestDocument(t, markup)
	page := newMemoryPage(t, `<html><body><div>page</div></body></html>`, drivers.NewHTTPCookies())

	return &testPage{
		HTMLPage: page,
		frame:    frame,
	}
}

func newTestDocument(t *testing.T, markup string) *testDocument {
	t.Helper()

	base := newMemoryDocument(t, markup)
	element, ok := base.GetElement().(*memory.HTMLElement)
	if !ok {
		t.Fatalf("expected memory element, got %T", base.GetElement())
	}

	return &testDocument{
		HTMLDocument: base,
		element: &testElement{
			HTMLElement: element,
		},
	}
}

func newMemoryPage(t *testing.T, markup string, cookies *drivers.HTTPCookies) *memory.HTMLPage {
	t.Helper()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(markup))
	if err != nil {
		t.Fatalf("failed to parse document: %v", err)
	}

	page, err := memory.NewHTMLPage(
		doc,
		"https://example.com",
		drivers.HTTPResponse{Headers: drivers.NewHTTPHeaders()},
		cookies,
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
