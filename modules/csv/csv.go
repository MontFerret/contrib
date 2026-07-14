package csv

import (
	"github.com/MontFerret/contrib/modules/csv/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the CSV module, which registers the CSV namespace functions on a
// Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("csv", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("CSV"))
	})
}
