package lib

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func RegisterLib(ns runtime.Namespace) error {
	definitions := append(
		coreFunctionDefinitions(),
		sdk.Func("OPEN", Open),
		sdk.Func("CAN_OPEN", PageExists),
	)

	return sdk.RegisterFunctions(ns, definitions...)
}

func RegisterLibLegacy(ns runtime.Namespace) error {
	definitions := append(
		coreFunctionDefinitions(),
		sdk.Func("DOCUMENT", Open),
		sdk.Func("DOCUMENT_EXISTS", PageExists),
	)

	return sdk.RegisterFunctions(ns, definitions...)
}

func coreFunctionDefinitions() []sdk.FunctionDef {
	return []sdk.FunctionDef{
		sdk.Func("IS_HTML_ELEMENT", IsHTMLElement),
		sdk.Func("IS_HTML_DOCUMENT", IsHTMLDocument),
		sdk.Func("X", XPathSelector),
		sdk.Func("XPATH", XPath),
		sdk.Func("FRAMES", Frames),
		sdk.Func("ATTR_GET", AttributeGet),
		sdk.Func("ATTR_QUERY", AttributeQuery),
		sdk.Func("ATTR_REMOVE", AttributeRemove),
		sdk.Func("ATTR_SET", AttributeSet),
		sdk.Func("BLUR", Blur),
		sdk.Func("COOKIE_DEL", CookieDel),
		sdk.Func("COOKIE_GET", CookieGet),
		sdk.Func("COOKIE_SET", CookieSet),
		sdk.Func("CLICK", Click),
		sdk.Func("CLICK_ALL", ClickAll),
		sdk.Func("DOWNLOAD", Download),
		sdk.Func("ELEMENT", Element),
		sdk.Func("ELEMENT_EXISTS", ElementExists),
		sdk.Func("ELEMENTS", Elements),
		sdk.Func("ELEMENTS_COUNT", ElementsCount),
		sdk.Func("FOCUS", Focus),
		sdk.Func("HOVER", Hover),
		sdk.Func("INNER_HTML", GetInnerHTML),
		sdk.Func("INNER_HTML_SET", SetInnerHTML),
		sdk.Func("INNER_HTML_ALL", GetInnerHTMLAll),
		sdk.Func("INNER_TEXT", GetInnerText),
		sdk.Func("INNER_TEXT_SET", SetInnerText),
		sdk.Func("INNER_TEXT_ALL", GetInnerTextAll),
		sdk.Func("INPUT", Input),
		sdk.Func("INPUT_CLEAR", InputClear),
		sdk.Func("MOUSE", MouseMoveXY),
		sdk.Func("NAVIGATE", Navigate),
		sdk.Func("NAVIGATE_BACK", NavigateBack),
		sdk.Func("NAVIGATE_FORWARD", NavigateForward),
		sdk.Func("PAGINATION", Pagination),
		sdk.Func("PARSE", Parse),
		sdk.Func("PDF", PDF),
		sdk.Func("PRESS", Press),
		sdk.Func("PRESS_SELECTOR", PressSelector),
		sdk.Func("SCREENSHOT", Screenshot),
		sdk.Func("SCROLL", ScrollXY),
		sdk.Func("SCROLL_BOTTOM", ScrollBottom),
		sdk.Func("SCROLL_ELEMENT", ScrollInto),
		sdk.Func("SCROLL_TOP", ScrollTop),
		sdk.Func("SELECT", Select),
		sdk.Func("STYLE_GET", StyleGet),
		sdk.Func("STYLE_REMOVE", StyleRemove),
		sdk.Func("STYLE_SET", StyleSet),
		sdk.Func("WAIT_ATTR", WaitAttribute),
		sdk.Func("WAIT_NO_ATTR", WaitNoAttribute),
		sdk.Func("WAIT_ATTR_ALL", WaitAttributeAll),
		sdk.Func("WAIT_NO_ATTR_ALL", WaitNoAttributeAll),
		sdk.Func("WAIT_ELEMENT", WaitElement),
		sdk.Func("WAIT_NO_ELEMENT", WaitNoElement),
		sdk.Func("WAIT_CLASS", WaitClass),
		sdk.Func("WAIT_NO_CLASS", WaitNoClass),
		sdk.Func("WAIT_CLASS_ALL", WaitClassAll),
		sdk.Func("WAIT_NO_CLASS_ALL", WaitNoClassAll),
		sdk.Func("WAIT_STYLE", WaitStyle),
		sdk.Func("WAIT_NO_STYLE", WaitNoStyle),
		sdk.Func("WAIT_STYLE_ALL", WaitStyleAll),
		sdk.Func("WAIT_NO_STYLE_ALL", WaitNoStyleAll),
		sdk.Func("WAIT_NAVIGATION", WaitNavigation),
	}
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

func toScrollOptions(ctx context.Context, value runtime.Value) (drivers.ScrollOptions, error) {
	result := drivers.ScrollOptions{}

	if err := sdk.Decode(ctx, value, &result, sdk.DisallowUnknownFields()); err != nil {
		return result, err
	}

	return result, nil
}
