package common

import (
	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func CastHTMLDocument(value runtime.Value) (drivers.HTMLDocument, error) {
	doc, ok := value.(drivers.HTMLDocument)

	if !ok {
		return nil, runtime.TypeErrorOf(value, drivers.HTMLDocumentType)
	}

	return doc, nil
}

func CastHTMLElement(value runtime.Value) (drivers.HTMLElement, error) {
	node, ok := value.(drivers.HTMLElement)

	if !ok {
		return nil, runtime.TypeErrorOf(value, drivers.HTMLElementType)
	}

	return node, nil
}
