package pdf

import (
	"github.com/MontFerret/contrib/modules/document/pdf/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the DOCUMENT::PDF module, which registers read-only PDF helpers
// on a Ferret host during bootstrap.
func New(opts ...Option) module.Module {
	o := newOptions(opts)

	return sdk.NewModule("document/pdf", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(
			bootstrap.Host().Library().Namespace("DOCUMENT").Namespace("PDF"),
			o.openOptions,
		)
	})
}
