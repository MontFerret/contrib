package lib

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func RegisterLib(ns runtime.Namespace) {
	ns.Function().A1().
		Add("IS_HTML_ELEMENT", IsHTMLElement).
		Add("IS_HTML_DOCUMENT", IsHTMLDocument).
		Add("X", XPathSelector)

	ns.Function().A2().
		Add("XPATH", XPath)

	ns.Function().Var().
		Add("ATTR_GET", AttributeGet).
		Add("ATTR_QUERY", AttributeQuery).
		Add("ATTR_REMOVE", AttributeRemove).
		Add("ATTR_SET", AttributeSet).
		Add("BLUR", Blur).
		Add("COOKIE_DEL", CookieDel).
		Add("COOKIE_GET", CookieGet).
		Add("COOKIE_SET", CookieSet).
		Add("CLICK", Click).
		Add("CLICK_ALL", ClickAll).
		Add("DOCUMENT", Open).
		Add("DOCUMENT_EXISTS", DocumentExists).
		Add("DOWNLOAD", Download).
		Add("ELEMENT", Element).
		Add("ELEMENT_EXISTS", ElementExists).
		Add("ELEMENTS", Elements).
		Add("ELEMENTS_COUNT", ElementsCount).
		Add("FRAMES", Frames).
		Add("FOCUS", Focus).
		Add("HOVER", Hover).
		Add("INNER_HTML", GetInnerHTML).
		Add("INNER_HTML_SET", SetInnerHTML).
		Add("INNER_HTML_ALL", GetInnerHTMLAll).
		Add("INNER_TEXT", GetInnerText).
		Add("INNER_TEXT_SET", SetInnerText).
		Add("INNER_TEXT_ALL", GetInnerTextAll).
		Add("INPUT", Input).
		Add("INPUT_CLEAR", InputClear).
		Add("MOUSE", MouseMoveXY).
		Add("NAVIGATE", Navigate).
		Add("NAVIGATE_BACK", NavigateBack).
		Add("NAVIGATE_FORWARD", NavigateForward).
		Add("PAGINATION", Pagination).
		Add("PARSE", Parse).
		Add("PDF", PDF).
		Add("PRESS", Press).
		Add("PRESS_SELECTOR", PressSelector).
		Add("SCREENSHOT", Screenshot).
		Add("SCROLL", ScrollXY).
		Add("SCROLL_BOTTOM", ScrollBottom).
		Add("SCROLL_ELEMENT", ScrollInto).
		Add("SCROLL_TOP", ScrollTop).
		Add("SELECT", Select).
		Add("STYLE_GET", StyleGet).
		Add("STYLE_REMOVE", StyleRemove).
		Add("STYLE_SET", StyleSet).
		Add("WAIT_ATTR", WaitAttribute).
		Add("WAIT_NO_ATTR", WaitNoAttribute).
		Add("WAIT_ATTR_ALL", WaitAttributeAll).
		Add("WAIT_NO_ATTR_ALL", WaitNoAttributeAll).
		Add("WAIT_ELEMENT", WaitElement).
		Add("WAIT_NO_ELEMENT", WaitNoElement).
		Add("WAIT_CLASS", WaitClass).
		Add("WAIT_NO_CLASS", WaitNoClass).
		Add("WAIT_CLASS_ALL", WaitClassAll).
		Add("WAIT_NO_CLASS_ALL", WaitNoClassAll).
		Add("WAIT_STYLE", WaitStyle).
		Add("WAIT_NO_STYLE", WaitNoStyle).
		Add("WAIT_STYLE_ALL", WaitStyleAll).
		Add("WAIT_NO_STYLE_ALL", WaitNoStyleAll).
		Add("WAIT_NAVIGATION", WaitNavigation)
}

func OpenOrCastPage(ctx context.Context, value runtime.Value) (drivers.HTMLPage, bool, error) {
	var page drivers.HTMLPage
	var closeAfter bool

	switch argv := value.(type) {
	case runtime.String:
		buf, err := Open(ctx, value, runtime.NewBoolean(true))

		if err != nil {
			return nil, false, err
		}

		page = buf.(drivers.HTMLPage)
		closeAfter = true
	case drivers.HTMLPage:
		page = argv
	default:
		return nil, false, runtime.TypeError(runtime.TypeOf(value), drivers.HTMLPageType, runtime.TypeString)
	}

	return page, closeAfter, nil
}

func waitTimeout(ctx context.Context, value runtime.Int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(
		ctx,
		time.Duration(value)*time.Millisecond,
	)
}

func toScrollOptions(value runtime.Value) (drivers.ScrollOptions, error) {
	result := drivers.ScrollOptions{}

	if err := sdk.Decode(value, &result); err != nil {
		return result, err
	}

	return result, nil
}
