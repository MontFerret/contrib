package jwt

import (
	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/contrib/modules/security/jwt/lib"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
	config core.Config
}

// New returns the SECURITY::JWT module, which registers JWT helpers on a Ferret host during bootstrap.
func New(opts ...Option) module.Module {
	o := newOptions(opts)

	return &mod{
		config: o.coreConfig(),
	}
}

func (m *mod) Name() string {
	return "security/jwt"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	lib.RegisterLib(
		bootstrap.Host().Library().Namespace("SECURITY").Namespace("JWT"),
		m.config,
	)

	return nil
}
