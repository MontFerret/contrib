package drivers_test

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type providerWaitTarget struct{}

func (target *providerWaitTarget) WaitForElement(context.Context, drivers.QuerySelector, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForElementAll(context.Context, drivers.QuerySelector, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForAttribute(context.Context, runtime.String, runtime.Value, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForAttributeBySelector(context.Context, drivers.QuerySelector, runtime.String, runtime.Value, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForAttributeBySelectorAll(context.Context, drivers.QuerySelector, runtime.String, runtime.Value, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForStyle(context.Context, runtime.String, runtime.Value, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForStyleBySelector(context.Context, drivers.QuerySelector, runtime.String, runtime.Value, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForStyleBySelectorAll(context.Context, drivers.QuerySelector, runtime.String, runtime.Value, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForClass(context.Context, runtime.String, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForClassBySelector(context.Context, drivers.QuerySelector, runtime.String, drivers.WaitEvent) error {
	return nil
}

func (target *providerWaitTarget) WaitForClassBySelectorAll(context.Context, drivers.QuerySelector, runtime.String, drivers.WaitEvent) error {
	return nil
}
