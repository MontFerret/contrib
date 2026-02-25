package html

import (
	"context"
	"time"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func RegisterLib(ns runtime.Namespace) error {
	return ns.RegisterFunctions(
		runtime.NewFunctionsFromMap(map[string]runtime.Function{
			"ATTR_GET":          AttributeGet,
			"ATTR_QUERY":        AttributeQuery,
			"ATTR_REMOVE":       AttributeRemove,
			"ATTR_SET":          AttributeSet,
			"BLUR":              Blur,
			"COOKIE_DEL":        CookieDel,
			"COOKIE_GET":        CookieGet,
			"COOKIE_SET":        CookieSet,
			"CLICK":             Click,
			"CLICK_ALL":         ClickAll,
			"DOCUMENT":          Open,
			"DOCUMENT_EXISTS":   DocumentExists,
			"DOWNLOAD":          Download,
			"ELEMENT":           Element,
			"ELEMENT_EXISTS":    ElementExists,
			"ELEMENTS":          Elements,
			"ELEMENTS_COUNT":    ElementsCount,
			"FRAMES":            Frames,
			"FOCUS":             Focus,
			"HOVER":             Hover,
			"INNER_HTML":        GetInnerHTML,
			"INNER_HTML_SET":    SetInnerHTML,
			"INNER_HTML_ALL":    GetInnerHTMLAll,
			"INNER_TEXT":        GetInnerText,
			"INNER_TEXT_SET":    SetInnerText,
			"INNER_TEXT_ALL":    GetInnerTextAll,
			"INPUT":             Input,
			"INPUT_CLEAR":       InputClear,
			"MOUSE":             MouseMoveXY,
			"NAVIGATE":          Navigate,
			"NAVIGATE_BACK":     NavigateBack,
			"NAVIGATE_FORWARD":  NavigateForward,
			"PAGINATION":        Pagination,
			"PARSE":             Parse,
			"PDF":               PDF,
			"PRESS":             Press,
			"PRESS_SELECTOR":    PressSelector,
			"SCREENSHOT":        Screenshot,
			"SCROLL":            ScrollXY,
			"SCROLL_BOTTOM":     ScrollBottom,
			"SCROLL_ELEMENT":    ScrollInto,
			"SCROLL_TOP":        ScrollTop,
			"SELECT":            Select,
			"STYLE_GET":         StyleGet,
			"STYLE_REMOVE":      StyleRemove,
			"STYLE_SET":         StyleSet,
			"WAIT_ATTR":         WaitAttribute,
			"WAIT_NO_ATTR":      WaitNoAttribute,
			"WAIT_ATTR_ALL":     WaitAttributeAll,
			"WAIT_NO_ATTR_ALL":  WaitNoAttributeAll,
			"WAIT_ELEMENT":      WaitElement,
			"WAIT_NO_ELEMENT":   WaitNoElement,
			"WAIT_CLASS":        WaitClass,
			"WAIT_NO_CLASS":     WaitNoClass,
			"WAIT_CLASS_ALL":    WaitClassAll,
			"WAIT_NO_CLASS_ALL": WaitNoClassAll,
			"WAIT_STYLE":        WaitStyle,
			"WAIT_NO_STYLE":     WaitNoStyle,
			"WAIT_STYLE_ALL":    WaitStyleAll,
			"WAIT_NO_STYLE_ALL": WaitNoStyleAll,
			"WAIT_NAVIGATION":   WaitNavigation,
			//"XPATH":             XPath,
			//"X":                 XPathSelector,
		}))
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

	// TODO: Add validation for the options
	return result, nil

	//err := runtime.ValidateType(value, runtime.TypeObject)
	//
	//if err != nil {
	//	return result, err
	//}
	//
	//obj := value.(*runtime.Object)
	//
	//behavior, exists := obj.Get("behavior")
	//
	//if exists {
	//	result.Behavior = drivers.NewScrollBehavior(behavior.String())
	//}
	//
	//block, exists := obj.Get("block")
	//
	//if exists {
	//	result.Block = drivers.NewScrollVerticalAlignment(block.String())
	//}
	//
	//inline, exists := obj.Get("inline")
	//
	//if exists {
	//	result.Inline = drivers.NewScrollHorizontalAlignment(inline.String())
	//}
}
