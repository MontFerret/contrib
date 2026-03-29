package html

import (
	"fmt"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

type (
	options struct {
		defaultDrv string
		drivers    []drivers.Driver
		noLib      bool
	}

	Option func(opts *options) error
)

func newOptions(opts []Option) (*options, error) {
	o := &options{}

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	return o, nil
}

func WithNoLib() Option {
	return func(o *options) error {
		o.noLib = true

		return nil
	}
}

func WithDefaultDriver(drv drivers.Driver) Option {
	return func(o *options) error {
		if drv == nil {
			return fmt.Errorf("driver cannot be nil")
		}

		if o.drivers == nil {
			o.drivers = make([]drivers.Driver, 0, 1)
		}

		o.drivers = append(o.drivers, drv)
		o.defaultDrv = drv.Name()

		return nil
	}
}

func WithDrivers(drvs ...drivers.Driver) Option {
	return func(o *options) error {
		if len(drvs) == 0 {
			return fmt.Errorf("drivers cannot be empty")
		}

		if o.drivers == nil {
			o.drivers = make([]drivers.Driver, 0, len(drvs))
		}

		for _, d := range drvs {
			if d == nil {
				return fmt.Errorf("driver cannot be nil")
			}

			o.drivers = append(o.drivers, d)
		}

		return nil
	}
}
