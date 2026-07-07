package xlsx

import (
	"github.com/MontFerret/contrib/modules/document/xlsx/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct{}

// New returns the DOCUMENT::XLSX module, which registers Excel workbook helpers
// on a Ferret host during bootstrap.
func New() module.Module {
	return &mod{}
}

func (m *mod) Name() string {
	return "document/xlsx"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("DOCUMENT").Namespace("XLSX"))

	return nil
}
