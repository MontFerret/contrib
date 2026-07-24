package oauth2

import (
	"github.com/MontFerret/contrib/modules/security/oauth2/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the SECURITY::OAUTH2 module, which registers OAuth 2.0 client
// helpers on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("security/oauth2", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(
			bootstrap.Host().Library().Namespace("SECURITY").Namespace("OAUTH2"),
		)
	})
}
