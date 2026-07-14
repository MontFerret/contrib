package xlsx

import (
	"github.com/MontFerret/contrib/modules/document/xlsx/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the DOCUMENT::XLSX module, which registers Excel workbook helpers
// on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("document/xlsx", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("DOCUMENT").Namespace("XLSX"))
	})
}
