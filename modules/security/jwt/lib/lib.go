package lib

import (
	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// RegisterLib registers SECURITY::JWT namespace functions.
func RegisterLib(ns runtime.Namespace, cfg core.Config) {
	ns.Function().Var().
		Add("INSPECT", inspectWithConfig(cfg)).
		Add("VERIFY", verifyWithConfig(cfg)).
		Add("SIGN", signWithConfig(cfg)).
		Add("HMAC_KEY", HMACKey).
		Add("PUBLIC_KEY", PublicKey).
		Add("PRIVATE_KEY", PrivateKey)
}
