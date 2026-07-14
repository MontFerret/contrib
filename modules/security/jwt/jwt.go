package jwt

import (
	"github.com/MontFerret/contrib/modules/security/jwt/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the SECURITY::JWT module, which registers JWT helpers on a Ferret host during bootstrap.
func New(opts ...Option) module.Module {
	o := newOptions(opts)
	config := o.coreConfig()

	return sdk.NewModule("security/jwt", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(
			bootstrap.Host().Library().Namespace("SECURITY").Namespace("JWT"),
			config,
		)
	})
}
