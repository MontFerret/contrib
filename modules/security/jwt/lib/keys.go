package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/security/jwt/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// HMACKey creates an opaque HMAC key from secret material.
func HMACKey(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	secret, err := core.ResolveSecret(args[0])
	if err != nil {
		return nil, core.OperationError("HMAC_KEY", err)
	}

	if len(secret) == 0 {
		return nil, core.OperationError("HMAC_KEY", core.NewInvalidKeyError("HMAC secret must not be empty"))
	}

	return core.NewHMACKey(secret), nil
}

// PublicKey parses a PEM-encoded public key into an opaque key handle.
func PublicKey(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	pemText, err := runtime.CastArgAt[runtime.String](args, 0)
	if err != nil {
		return nil, core.OperationError("PUBLIC_KEY", err)
	}

	key, err := core.NewPublicKey(pemText.String())
	if err != nil {
		return nil, core.OperationError("PUBLIC_KEY", err)
	}

	return key, nil
}

// PrivateKey parses a PEM-encoded private key into an opaque key handle.
func PrivateKey(ctx context.Context, args ...runtime.Value) (runtime.Value, error) {
	if err := runtime.ValidateArgs(args, 1, 1); err != nil {
		return nil, err
	}

	pemText, err := runtime.CastArgAt[runtime.String](args, 0)
	if err != nil {
		return nil, core.OperationError("PRIVATE_KEY", err)
	}

	key, err := core.NewPrivateKey(pemText.String())
	if err != nil {
		return nil, core.OperationError("PRIVATE_KEY", err)
	}

	return key, nil
}
