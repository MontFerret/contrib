package archive

import (
	"context"
	"fmt"

	"github.com/ziflex/go-options"

	"github.com/MontFerret/contrib/modules/archive/core"
	"github.com/MontFerret/contrib/modules/archive/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the ARCHIVE module.
func New(setters ...Option) module.Module {
	return sdk.NewModule("archive", func(bootstrap module.Bootstrap) error {
		config, err := options.ApplyWithValues[core.Config](core.DefaultConfig(), setters...)
		if err != nil {
			return fmt.Errorf("configure archive module: %w", err)
		}

		if err := lib.RegisterLib(bootstrap.Host().Library().Namespace("ARCHIVE")); err != nil {
			return err
		}

		bootstrap.Hooks().Session().BeforeRun(func(ctx context.Context) (context.Context, error) {
			return core.WithConfig(ctx, config), nil
		})

		return nil
	})
}
