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
	return toRootElementCapability[drivers.ContentTarget](value, "content")
}

func toRootAttributeTarget(value runtime.Value) (drivers.AttributeTarget, error) {
	return toRootElementCapability[drivers.AttributeTarget](value, "attribute")
}

func toRootInteractionTarget(value runtime.Value) (drivers.InteractionTarget, error) {
	return toRootElementCapability[drivers.InteractionTarget](value, "interaction")
}

func toRootWaitTarget(value runtime.Value) (drivers.WaitTarget, error) {
	return toRootElementCapability[drivers.WaitTarget](value, "wait")
}

func toRootElementCapability[T any](value runtime.Value, capability string) (T, error) {
	var zero T

	el, err := toRootElement(value)
	if err != nil {
		return zero, err
	}

	target, ok := any(el).(T)
	if !ok {
		return zero, runtime.Errorf(runtime.ErrNotSupported, "root element %s capability", capability)
	}

	return target, nil
}
