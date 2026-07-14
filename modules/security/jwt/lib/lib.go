package lib

import (
	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/pkg/sdk"
)

// RegisterLib registers SECURITY::JWT namespace functions.
func RegisterLib(ns runtime.Namespace, cfg core.Config) error {
	return sdk.RegisterFunctions(
		ns,
		sdk.Func("INSPECT", inspectWithConfig(cfg)),
		sdk.Func("VERIFY", verifyWithConfig(cfg)),
		sdk.Func("SIGN", signWithConfig(cfg)),
		sdk.Func("HMAC_KEY", HMACKey),
		sdk.Func("PUBLIC_KEY", PublicKey),
		sdk.Func("PRIVATE_KEY", PrivateKey),
	)
}
