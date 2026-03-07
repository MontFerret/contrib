package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/contrib/modules/html/lib"
	"github.com/MontFerret/ferret/v2"
)

type module struct {
	drivers *drivers.Container
	noLib   bool
}

func New(opts ...Option) (ferret.Module, error) {
	o, err := newOptions(opts)

	if err != nil {
		return nil, err
	}

	container := drivers.NewContainer()

	for _, d := range o.drivers {
		if err := container.Register(d, o.globalOpts...); err != nil {
			return nil, err
		}
	}

	return &module{
		drivers: container,
		noLib:   o.noLib,
	}, nil
}

func (m *module) Name() string {
	return "html"
}

func (m *module) Register(registry ferret.Bootstrap) error {
	if !m.noLib {
		lib.RegisterLib(registry.Host().Library())
	}

	registry.Hooks().Session().BeforeRun(func(ctx context.Context) (context.Context, error) {
		return m.drivers.WithContext(ctx), nil
	})

	return nil
}
