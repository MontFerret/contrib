package drivers

import (
	"context"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type Container struct {
	drivers    map[string]Driver
	defaultDrv string
}

func NewContainer() *Container {
	return &Container{
		drivers: map[string]Driver{},
	}
}

func (c *Container) Has(name string) bool {
	_, exists := c.drivers[name]

	return exists
}

func (c *Container) Default() (Driver, bool) {
	if c.defaultDrv == "" {
		return nil, false
	}

	return c.Get(c.defaultDrv)
}

func (c *Container) SetDefault(name string) {
	c.defaultDrv = name
}

func (c *Container) Register(drv Driver) error {
	if drv == nil {
		return runtime.Error(runtime.ErrMissedArgument, "driver")
	}

	name := drv.Name()
	_, exists := c.drivers[name]

	if exists {
		return runtime.Errorf(runtime.ErrNotUnique, "driver: %s", name)
	}

	c.drivers[name] = drv

	return nil
}

func (c *Container) Remove(name string) {
	delete(c.drivers, name)
}

func (c *Container) Get(name string) (Driver, bool) {
	if name == "" {
		name = c.defaultDrv
	}

	found, exists := c.drivers[name]

	return found, exists
}

func (c *Container) GetAll() []Driver {
	res := make([]Driver, 0, len(c.drivers))

	for _, drv := range c.drivers {
		res = append(res, drv)
	}

	return res
}

func (c *Container) WithContext(ctx context.Context) context.Context {
	key := ctxKey{}

	return context.WithValue(ctx, key, c)
}
