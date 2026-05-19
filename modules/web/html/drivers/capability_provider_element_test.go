package drivers_test

import "github.com/MontFerret/contrib/modules/web/html/drivers"

type providerElement struct {
	drivers.HTMLNode
	attrs  drivers.AttributeTarget
	styles drivers.StyleTarget
	wait   drivers.WaitTarget
}

func (el *providerElement) AsAttributeTarget() drivers.AttributeTarget {
	return el.attrs
}

func (el *providerElement) AsStyleTarget() drivers.StyleTarget {
	return el.styles
}

func (el *providerElement) AsWaitTarget() drivers.WaitTarget {
	return el.wait
}
