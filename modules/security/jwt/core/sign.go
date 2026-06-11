package core

import (
	"context"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Sign creates a compact signed JWT.
func Sign(ctx context.Context, claimsMap runtime.Map, key runtime.Value, opts SignOptions) (runtime.Value, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	claims, err := runtimeMapToClaims(ctx, claimsMap)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	if err := detectClaimConflicts(claims, opts, now); err != nil {
		return nil, err
	}

	if err := applySignOptions(claims, opts, now); err != nil {
		return nil, err
	}

	method, err := signingMethod(opts.Algorithm)
	if err != nil {
		return nil, err
	}

	keyMaterial, err := signKeyMaterial(key, opts.Algorithm)
	if err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(method, jwt.MapClaims(claims))
	token.Header["alg"] = opts.Algorithm
	token.Header["typ"] = "JWT"

	for key, value := range opts.Header {
		if key == "alg" {
			continue
		}

		token.Header[key] = value
	}

	compact, err := token.SignedString(keyMaterial)
	if err != nil {
		return nil, wrapError(ErrInvalidToken, "failed to sign JWT", err)
	}

	return runtime.NewString(compact), nil
}

func detectClaimConflicts(claims map[string]any, opts SignOptions, now int64) error {
	if opts.Issuer != "" && hasClaim(claims, "iss") {
		return newError(ErrInvalidToken, "issuer conflicts with claims")
	}

	if opts.Audience != nil && hasClaim(claims, "aud") {
		return newError(ErrInvalidToken, "audience conflicts with claims")
	}

	if opts.Subject != "" && hasClaim(claims, "sub") {
		return newError(ErrInvalidToken, "subject conflicts with claims")
	}

	if opts.ExpiresIn > 0 && hasClaim(claims, "exp") {
		return newError(ErrInvalidToken, "expires_in conflicts with claims")
	}

	if opts.NotBefore != 0 && hasClaim(claims, "nbf") {
		return newError(ErrInvalidToken, "not_before conflicts with claims")
	}

	if opts.IssuedAt && hasClaim(claims, "iat") {
		return newError(ErrInvalidToken, "issued_at conflicts with claims")
	}

	return nil
}

func applySignOptions(claims map[string]any, opts SignOptions, now int64) error {
	if opts.Issuer != "" {
		claims["iss"] = opts.Issuer
	}

	if opts.Audience != nil {
		claims["aud"] = opts.Audience
	}

	if opts.Subject != "" {
		claims["sub"] = opts.Subject
	}

	if opts.ExpiresIn > 0 {
		claims["exp"] = now + opts.ExpiresIn
	}

	if opts.NotBefore != 0 {
		claims["nbf"] = opts.NotBefore
	}

	if opts.IssuedAt {
		claims["iat"] = now
	}

	return nil
}
