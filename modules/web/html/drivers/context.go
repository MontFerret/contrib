package drivers

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type ctxKey struct{}

func FromContext(ctx context.Context, name string) (Driver, error) {
	container := resolveValue(ctx)

	if container == nil {
		return nil, runtime.Error(runtime.ErrNotFound, "make sure the module is registered")
	}

	if name == "" {
		def, exists := container.Default()

		if !exists {
			return nil, runtime.Error(runtime.ErrNotFound, "default driver is not set")
		}

		return def, nil
	}

	drv, exists := container.Get(name)

	if !exists {
		drvNotFound := runtime.Errorf(runtime.ErrNotFound, "driver: %s", name)

		return nil, runtime.Error(drvNotFound, "make sure the driver is registered during module initialization")
	}

	return drv, nil
}

func resolveValue(ctx context.Context) *Container {
	key := ctxKey{}
	v := ctx.Value(key)
	value, ok := v.(*Container)

	if !ok {
		return nil
	}

	return value
}
