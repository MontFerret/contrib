package lib

import (
	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func toRootElement(value runtime.Value) (drivers.HTMLElement, error) {
	switch v := value.(type) {
	case drivers.HTMLPage:
		return drivers.ToElement(v.GetMainFrame().GetElement())
	case drivers.HTMLDocument:
		return drivers.ToElement(v.GetElement())
	case drivers.HTMLElement:
		return drivers.ToElement(v)
	default:
		return nil, runtime.TypeErrorOf(value, drivers.HTMLPageType, drivers.HTMLDocumentType, drivers.HTMLElementType)
	}
}

func toRootContentTarget(value runtime.Value) (drivers.ContentTarget, error) {
	el, err := toRootElement(value)
	if err != nil {
		return nil, err
	}

	return drivers.ToContentTarget(el)
}

func toRootAttributeTarget(value runtime.Value) (drivers.AttributeTarget, error) {
	el, err := toRootElement(value)
	if err != nil {
		return nil, err
	}

	return drivers.ToAttributeTarget(el)
}

func toRootInteractionTarget(value runtime.Value) (drivers.InteractionTarget, error) {
	el, err := toRootElement(value)
	if err != nil {
		return nil, err
	}

	return drivers.ToInteractionTarget(el)
}

func toRootWaitTarget(value runtime.Value) (drivers.WaitTarget, error) {
	el, err := toRootElement(value)
	if err != nil {
		return nil, err
	}

	return drivers.ToWaitTarget(el)
}
