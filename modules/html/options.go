package html

import (
	"fmt"

	"github.com/MontFerret/contrib/modules/html/drivers"
)

type (
	options struct {
		noLib      bool
		drivers    []drivers.Driver
		globalOpts []drivers.GlobalOption
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

func WithGlobalOptions(opts ...drivers.GlobalOption) Option {
	return func(o *options) error {
		if len(opts) == 0 {
			return fmt.Errorf("global options must not be empty")
		}

		if o.globalOpts == nil {
			o.globalOpts = make([]drivers.GlobalOption, 0, len(opts))
		}

		for _, opt := range opts {
			if opt == nil {
				return fmt.Errorf("global option cannot be nil")
			}

			o.globalOpts = append(o.globalOpts, opt)
		}

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
