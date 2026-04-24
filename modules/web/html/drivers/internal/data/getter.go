package data

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func GetInPage(ctx context.Context, key runtime.Value, page drivers.HTMLPage) (runtime.Value, error) {
	if isEmptyValue(key) {
		return runtime.None, nil
	}

	switch key.String() {
	case "response":
		target, ok := page.(drivers.PageResponseTarget)
		if !ok {
			return runtime.None, runtime.Errorf(runtime.ErrNotSupported, "page response capability")
		}

		resp, err := target.GetResponse(ctx)
		if err != nil {
			return nil, err
		}

		return &resp, nil
	case "mainFrame", "document":
		return page.GetMainFrame(), nil
	case "frames":
		return page.GetFrames(ctx)
	case "url", "URL":
		return page.GetURL(), nil
	case "cookies":
		target, ok := page.(drivers.PageCookieReader)
		if !ok {
			return runtime.None, runtime.Errorf(runtime.ErrNotSupported, "page cookies capability")
		}

		cookies, err := target.GetCookies(ctx)
		if err != nil {
			return nil, err
		}

		return drivers.NewHTTPCookiesFrom(cookies), nil
	case "title":
		return page.GetMainFrame().GetTitle(), nil
	case "isClosed":
		return page.IsClosed(), nil
	default:
		return GetInDocument(ctx, key, page.GetMainFrame())
	}
}

func GetInDocument(ctx context.Context, key runtime.Value, doc drivers.HTMLDocument) (runtime.Value, error) {
	if isEmptyValue(key) {
		return runtime.None, nil
	}

	switch key.String() {
	case "url", "URL":
		return doc.GetURL(), nil
	case "name":
		return doc.GetName(), nil
	case "title":
		return doc.GetTitle(), nil
	case "parent":
		return doc.GetParentDocument(ctx)
	case "body", "head":
		return doc.QuerySelector(ctx, drivers.NewCSSSelector(runtime.String(key.String())))
	case "innerHTML":
		return doc.GetElement().GetInnerHTML(ctx)
	case "innerText":
		return doc.GetElement().GetInnerText(ctx)
	default:
		return GetInNode(ctx, key, doc.GetElement())
	}
}

func GetInElement(ctx context.Context, key runtime.Value, el drivers.HTMLElement) (runtime.Value, error) {
	if isEmptyValue(key) {
		return runtime.None, nil
	}

	switch key.String() {
	case "innerText":
		return el.GetInnerText(ctx)
	case "innerHTML":
		return el.GetInnerHTML(ctx)
	case "value":
		return el.GetValue(ctx)
	case "attributes":
		return el.GetAttributes(ctx)
	case "style":
		return el.GetStyles(ctx)
	case "previousElementSibling":
		return el.GetPreviousElementSibling(ctx)
	case "nextElementSibling":
		return el.GetNextElementSibling(ctx)
	case "parentElement":
		return el.GetParentElement(ctx)
	default:
		return GetInNode(ctx, key, el)
	}
}

func GetInNode(ctx context.Context, key runtime.Value, node drivers.HTMLNode) (runtime.Value, error) {
	if isEmptyValue(key) {
		return runtime.None, nil
	}

	switch keyVal := key.(type) {
	case runtime.Int:
		return node.GetChildNode(ctx, keyVal)
	case runtime.String:
		switch keyVal {
		case "nodeType":
			return node.GetNodeType(ctx)
		case "nodeName":
			return node.GetNodeName(ctx)
		case "children":
			return node.GetChildNodes(ctx)
		case "length":
			return node.Length(ctx)
		default:
			return runtime.None, nil
		}
	default:
		return runtime.None, nil
	}
}

func isEmptyValue(value runtime.Value) bool {
	if value == nil {
		return true
	}

	return value == runtime.None
}
