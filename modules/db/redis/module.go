package redis

import (
	"github.com/MontFerret/contrib/modules/db/redis/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// New returns the DB::REDIS module, which registers lifecycle helpers for Redis
// connection handles on a Ferret host during bootstrap.
func New() module.Module {
	return sdk.NewModule("db/redis", func(bootstrap module.Bootstrap) error {
		return lib.RegisterLib(bootstrap.Host().Library().Namespace("DB").Namespace("REDIS"))
	})
}
