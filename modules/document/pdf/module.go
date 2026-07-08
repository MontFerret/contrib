package pdf

import (
	"github.com/MontFerret/contrib/modules/document/pdf/core"
	"github.com/MontFerret/contrib/modules/document/pdf/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
	openOptions core.OpenOptions
}

// New returns the DOCUMENT::PDF module, which registers read-only PDF helpers
// on a Ferret host during bootstrap.
func New(opts ...Option) module.Module {
	o := newOptions(opts)

	return &mod{
		openOptions: o.openOptions,
	}
}

func (m *mod) Name() string {
	return "document/pdf"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(bootstrap.Host().Library().Namespace("DOCUMENT").Namespace("PDF"), m.openOptions)

	return nil
}
