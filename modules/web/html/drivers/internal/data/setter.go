package data

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func SetInPage(ctx context.Context, key runtime.Value, value runtime.Value, page drivers.HTMLPage) error {
	if isEmptyValue(key) {
		return nil
	}

	return SetInDocument(ctx, key, value, page.GetMainFrame())
}

func SetInDocument(ctx context.Context, key runtime.Value, value runtime.Value, doc drivers.HTMLDocument) error {
	if isEmptyValue(key) {
		return nil
	}

	return SetInElement(ctx, key, value, doc.GetElement())
}

func SetInElement(ctx context.Context, key runtime.Value, value runtime.Value, el drivers.HTMLElement) error {
	if isEmptyValue(key) {
		return nil
	}

	if value == nil {
		value = runtime.None
	}

	name := key.String()

	switch name {
	case "textContent":
		return el.SetTextContent(ctx, runtime.ToString(value))
	case "innerText":
		return el.SetInnerText(ctx, runtime.ToString(value))
	case "innerHTML":
		return el.SetInnerHTML(ctx, runtime.ToString(value))
	case "value":
		return el.SetValue(ctx, value)
	case "attributes":
		attrs, err := runtime.CastMap(value)
		if err != nil {
			return err
		}

		return el.SetAttributes(ctx, attrs)
	case "style":
		styles, err := runtime.CastMap(value)
		if err != nil {
			return err
		}

		return el.SetStyles(ctx, styles)
	default:
		return runtime.Errorf(runtime.ErrInvalidArgument, "element property %q is not writable", name)
	}
}
