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
		return valueOrNone(&resp, err)
	case "mainFrame", "document":
		return page.GetMainFrame(), nil
	case "frames":
		return valueOrNone(page.GetFrames(ctx))
	case "url", "URL":
		return page.GetURL(), nil
	case "cookies":
		target, ok := page.(drivers.PageCookieReader)
		if !ok {
			return runtime.None, runtime.Errorf(runtime.ErrNotSupported, "page cookies capability")
		}

		cookies, err := target.GetCookies(ctx)
		if err != nil {
			return runtime.None, err
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
		return valueOrNone(doc.GetParentDocument(ctx))
	case "body", "head":
		return valueOrNone(doc.QuerySelector(ctx, drivers.NewCSSSelector(runtime.String(key.String()))))
	case "innerHTML":
		target, err := drivers.ToContentTarget(doc.GetElement())
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetInnerHTML(ctx))
	case "innerText":
		target, err := drivers.ToContentTarget(doc.GetElement())
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetInnerText(ctx))
	default:
		return GetInNode(ctx, key, doc.GetElement())
	}
}

func GetInElement(ctx context.Context, key runtime.Value, el drivers.HTMLElement) (runtime.Value, error) {
	if isEmptyValue(key) {
		return runtime.None, nil
	}

	switch key.String() {
	case "textContent":
		target, err := drivers.ToContentTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetTextContent(ctx))
	case "innerText":
		target, err := drivers.ToContentTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetInnerText(ctx))
	case "innerHTML":
		target, err := drivers.ToContentTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetInnerHTML(ctx))
	case "value":
		target, err := drivers.ToValueTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetValue(ctx))
	case "attributes":
		target, err := drivers.ToAttributeTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetAttributes(ctx))
	case "style":
		target, err := drivers.ToStyleTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetStyles(ctx))
	case "previousElementSibling":
		target, err := drivers.ToRelationTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetPreviousElementSibling(ctx))
	case "nextElementSibling":
		target, err := drivers.ToRelationTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetNextElementSibling(ctx))
	case "parentElement":
		target, err := drivers.ToRelationTarget(el)
		if err != nil {
			return runtime.None, err
		}

		return valueOrNone(target.GetParentElement(ctx))
	default:
		value, err := GetInNode(ctx, key, el)
		if err != nil || value != runtime.None {
			return value, err
		}

		keyVal, ok := key.(runtime.String)
		if !ok {
			return runtime.None, nil
		}

		target, ok := el.(drivers.DOMPropertyTarget)
		if !ok {
			return runtime.None, nil
		}

		return valueOrNone(target.GetDOMProperty(ctx, keyVal))
	}
}

func GetInNode(ctx context.Context, key runtime.Value, node drivers.HTMLNode) (runtime.Value, error) {
	if isEmptyValue(key) {
		return runtime.None, nil
	}

	switch keyVal := key.(type) {
	case runtime.Int:
		return valueOrNone(node.GetChildNode(ctx, keyVal))
	case runtime.String:
		switch keyVal {
		case "nodeType":
			return valueOrNone(node.GetNodeType(ctx))
		case "nodeName":
			return valueOrNone(node.GetNodeName(ctx))
		case "children":
			return valueOrNone(node.GetChildNodes(ctx))
		case "length":
			return valueOrNone(node.Length(ctx))
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

func valueOrNone(value runtime.Value, err error) (runtime.Value, error) {
	if err != nil {
		return runtime.None, err
	}

	return value, nil
}
