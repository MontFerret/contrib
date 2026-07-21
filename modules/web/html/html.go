package html

import (
	"context"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

func New(opts ...Option) (module.Module, error) {
	o, err := newOptions(opts)

	if err != nil {
		return nil, err
	}

	container := drivers.NewContainer()

	for _, d := range o.drivers {
		if err := container.Register(d); err != nil {
			return nil, err
		}
	}

	if o.defaultDrv != "" {
		container.SetDefault(o.defaultDrv)
	}

	return sdk.NewModule("html", func(registry module.Bootstrap) error {
		if !o.noLib {
			// Legacy support for modules that use module functions without namespace (e.g. `DOCUMENT` instead of `WEB::HTML::DOCUMENT`).
			if err := lib.RegisterLibLegacy(registry.Host().Library()); err != nil {
				return err
			}
			if err := lib.RegisterLib(registry.Host().Library().Namespace("WEB").Namespace("HTML")); err != nil {
				return err
			}
		}

		registry.Hooks().Session().BeforeRun(func(ctx context.Context) (context.Context, error) {
			return container.WithContext(ctx), nil
		})

		return nil
	}), nil
}
