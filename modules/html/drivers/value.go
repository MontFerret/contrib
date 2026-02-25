package drivers

import (
	"context"
	"io"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	// WaitEvent is an enum that represents what event is needed to wait for
	WaitEvent int

	// HTMLNode is an interface from which a number of DOM API object types inherit.
	// It allows those types to be treated similarly;
	// for example, inheriting the same set of methods, or being tested in the same way.
	HTMLNode interface {
		runtime.Value
		runtime.Iterable
		runtime.KeyReadable
		runtime.KeyWritable
		runtime.Measurable
		runtime.Queryable
		runtime.Dispatchable
		io.Closer

		GetNodeType(ctx context.Context) (runtime.Int, error)

		GetNodeName(ctx context.Context) (runtime.String, error)

		GetChildNodes(ctx context.Context) (runtime.List, error)

		GetChildNode(ctx context.Context, idx runtime.Int) (runtime.Value, error)

		QuerySelector(ctx context.Context, selector QuerySelector) (runtime.Value, error)

		QuerySelectorAll(ctx context.Context, selector QuerySelector) (runtime.List, error)

		CountBySelector(ctx context.Context, selector QuerySelector) (runtime.Int, error)

		ExistsBySelector(ctx context.Context, selector QuerySelector) (runtime.Boolean, error)

		XPath(ctx context.Context, expression runtime.String) (runtime.Value, error)
	}

	// HTMLElement is the most general base interface which most objects in a GetMainFrame implement.
	HTMLElement interface {
		HTMLNode

		GetInnerText(ctx context.Context) (runtime.String, error)

		SetInnerText(ctx context.Context, innerText runtime.String) error

		GetInnerHTML(ctx context.Context) (runtime.String, error)

		SetInnerHTML(ctx context.Context, innerHTML runtime.String) error

		GetValue(ctx context.Context) (runtime.Value, error)

		SetValue(ctx context.Context, value runtime.Value) error

		GetStyles(ctx context.Context) (runtime.Map, error)

		GetStyle(ctx context.Context, name runtime.String) (runtime.Value, error)

		SetStyles(ctx context.Context, values runtime.Map) error

		SetStyle(ctx context.Context, name, value runtime.String) error

		RemoveStyle(ctx context.Context, name ...runtime.String) error

		GetAttributes(ctx context.Context) (runtime.Map, error)

		GetAttribute(ctx context.Context, name runtime.String) (runtime.Value, error)

		SetAttributes(ctx context.Context, values runtime.Map) error

		SetAttribute(ctx context.Context, name, value runtime.String) error

		RemoveAttribute(ctx context.Context, name ...runtime.String) error

		GetInnerHTMLBySelector(ctx context.Context, selector QuerySelector) (runtime.String, error)

		SetInnerHTMLBySelector(ctx context.Context, selector QuerySelector, innerHTML runtime.String) error

		GetInnerHTMLBySelectorAll(ctx context.Context, selector QuerySelector) (runtime.List, error)

		GetInnerTextBySelector(ctx context.Context, selector QuerySelector) (runtime.String, error)

		SetInnerTextBySelector(ctx context.Context, selector QuerySelector, innerText runtime.String) error

		GetInnerTextBySelectorAll(ctx context.Context, selector QuerySelector) (runtime.List, error)

		GetPreviousElementSibling(ctx context.Context) (runtime.Value, error)

		GetNextElementSibling(ctx context.Context) (runtime.Value, error)

		GetParentElement(ctx context.Context) (runtime.Value, error)

		Click(ctx context.Context, count runtime.Int) error

		ClickBySelector(ctx context.Context, selector QuerySelector, count runtime.Int) error

		ClickBySelectorAll(ctx context.Context, selector QuerySelector, count runtime.Int) error

		Clear(ctx context.Context) error

		ClearBySelector(ctx context.Context, selector QuerySelector) error

		Input(ctx context.Context, value runtime.Value, delay runtime.Int) error

		InputBySelector(ctx context.Context, selector QuerySelector, value runtime.Value, delay runtime.Int) error

		Press(ctx context.Context, keys []runtime.String, count runtime.Int) error

		PressBySelector(ctx context.Context, selector QuerySelector, keys []runtime.String, count runtime.Int) error

		Select(ctx context.Context, value runtime.List) (runtime.List, error)

		SelectBySelector(ctx context.Context, selector QuerySelector, value runtime.List) (runtime.List, error)

		ScrollIntoView(ctx context.Context, options ScrollOptions) error

		Focus(ctx context.Context) error

		FocusBySelector(ctx context.Context, selector QuerySelector) error

		Blur(ctx context.Context) error

		BlurBySelector(ctx context.Context, selector QuerySelector) error

		Hover(ctx context.Context) error

		HoverBySelector(ctx context.Context, selector QuerySelector) error

		WaitForElement(ctx context.Context, selector QuerySelector, when WaitEvent) error

		WaitForElementAll(ctx context.Context, selector QuerySelector, when WaitEvent) error

		WaitForAttribute(ctx context.Context, name runtime.String, value runtime.Value, when WaitEvent) error

		WaitForAttributeBySelector(ctx context.Context, selector QuerySelector, name runtime.String, value runtime.Value, when WaitEvent) error

		WaitForAttributeBySelectorAll(ctx context.Context, selector QuerySelector, name runtime.String, value runtime.Value, when WaitEvent) error

		WaitForStyle(ctx context.Context, name runtime.String, value runtime.Value, when WaitEvent) error

		WaitForStyleBySelector(ctx context.Context, selector QuerySelector, name runtime.String, value runtime.Value, when WaitEvent) error

		WaitForStyleBySelectorAll(ctx context.Context, selector QuerySelector, name runtime.String, value runtime.Value, when WaitEvent) error

		WaitForClass(ctx context.Context, class runtime.String, when WaitEvent) error

		WaitForClassBySelector(ctx context.Context, selector QuerySelector, class runtime.String, when WaitEvent) error

		WaitForClassBySelectorAll(ctx context.Context, selector QuerySelector, class runtime.String, when WaitEvent) error
	}

	HTMLDocument interface {
		HTMLNode

		GetTitle() runtime.String

		GetElement() HTMLElement

		GetURL() runtime.String

		GetName() runtime.String

		GetParentDocument(ctx context.Context) (HTMLDocument, error)

		GetChildDocuments(ctx context.Context) (runtime.List, error)

		Scroll(ctx context.Context, options ScrollOptions) error

		ScrollTop(ctx context.Context, options ScrollOptions) error

		ScrollBottom(ctx context.Context, options ScrollOptions) error

		ScrollBySelector(ctx context.Context, selector QuerySelector, options ScrollOptions) error

		MoveMouseByXY(ctx context.Context, x, y runtime.Float) error
	}

	// HTMLPage interface represents any web page loaded in the browser
	// and serves as an entry point into the web page's content
	HTMLPage interface {
		runtime.Value
		runtime.Iterable
		runtime.KeyReadable
		runtime.KeyWritable
		runtime.Measurable
		runtime.Observable
		runtime.Dispatchable
		io.Closer

		IsClosed() runtime.Boolean

		GetURL() runtime.String

		GetMainFrame() HTMLDocument

		GetFrames(ctx context.Context) (runtime.List, error)

		GetFrame(ctx context.Context, idx runtime.Int) (runtime.Value, error)

		GetCookies(ctx context.Context) (*HTTPCookies, error)

		SetCookies(ctx context.Context, cookies *HTTPCookies) error

		DeleteCookies(ctx context.Context, cookies *HTTPCookies) error

		GetResponse(ctx context.Context) (HTTPResponse, error)

		PrintToPDF(ctx context.Context, params PDFParams) (runtime.Binary, error)

		CaptureScreenshot(ctx context.Context, params ScreenshotParams) (runtime.Binary, error)

		WaitForNavigation(ctx context.Context, targetURL runtime.String) error

		WaitForFrameNavigation(ctx context.Context, frame HTMLDocument, targetURL runtime.String) error

		Navigate(ctx context.Context, url runtime.String) error

		NavigateBack(ctx context.Context, skip runtime.Int) (runtime.Boolean, error)

		NavigateForward(ctx context.Context, skip runtime.Int) (runtime.Boolean, error)
	}
)

const (
	// WaitEventPresence indicating to wait for Value to appear
	WaitEventPresence = 0

	// WaitEventAbsence indicating to wait for Value to disappear
	WaitEventAbsence = 1
)
