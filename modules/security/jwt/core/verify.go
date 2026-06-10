package core

import (
	"context"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Verify validates a compact JWT signature and claims.
func Verify(ctx context.Context, cfg Config, token runtime.String, key runtime.Value, opts VerifyOptions) (runtime.Value, error) {
	if err := opts.validate(); err != nil {
		return nil, err
	}

	parsed, err := parseCompactToken(token.String(), cfg.maxTokenSize())
	if err != nil {
		return nil, err
	}

	alg, err := headerAlgorithm(parsed.header)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(alg, "none") {
		return nil, newError(ErrUnexpectedAlgorithm, "alg none is not allowed")
	}

	allowed := normalizeAlgorithms(opts.Algorithms)
	if !algorithmAllowed(alg, allowed) {
		return nil, newError(ErrUnexpectedAlgorithm, "token algorithm is not allowed")
	}

	keyMaterial, err := verifyKeyMaterial(key, alg)
	if err != nil {
		return nil, err
	}

	parser := jwt.NewParser(jwt.WithValidMethods(allowed))
	jwtToken, err := parser.Parse(parsed.compact, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != alg {
			return nil, newError(ErrUnexpectedAlgorithm, "token algorithm is not allowed")
		}

		return keyMaterial, nil
	})
	if err != nil {
		if isSignatureError(err) {
			return nil, newError(ErrInvalidSignature, "invalid JWT signature")
		}

		return nil, wrapError(ErrInvalidToken, "invalid JWT token", err)
	}

	if !jwtToken.Valid {
		return nil, newError(ErrInvalidSignature, "invalid JWT signature")
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, newError(ErrInvalidToken, "invalid JWT claims")
	}

	normalized := mapClaimsToAny(claims)
	now := opts.now()

	if err := validateTimeClaims(normalized, now, opts.Leeway, opts.MaxAge); err != nil {
		return nil, err
	}

	if err := validateRegisteredClaims(normalized, opts); err != nil {
		return nil, err
	}

	parsed.claims = normalized

	return buildVerifyResult(parsed)
}

func isSignatureError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "signature")
}

func (o VerifyOptions) now() int64 {
	if o.Now > 0 {
		return o.Now
	}

	return time.Now().Unix()
}

func mapClaimsToAny(claims jwt.MapClaims) map[string]any {
	out := make(map[string]any, len(claims))
	for key, value := range claims {
		out[key] = value
	}

	return out
}

func validateTimeClaims(claims map[string]any, now int64, leeway int64, maxAge int64) error {
	if raw, ok := claims["exp"]; ok {
		exp, ok := claimAsInt64(raw)
		if !ok {
			return newError(ErrInvalidToken, "invalid exp claim")
		}

		if now > exp+leeway {
			return newError(ErrExpired, "token has expired")
		}
	}

	if raw, ok := claims["nbf"]; ok {
		nbf, ok := claimAsInt64(raw)
		if !ok {
			return newError(ErrInvalidToken, "invalid nbf claim")
		}

		if now+leeway < nbf {
			return newError(ErrNotYetValid, "token is not yet valid")
		}
	}

	if maxAge > 0 {
		raw, ok := claims["iat"]
		if !ok {
			return newError(ErrClaimMissing, "token is missing required iat claim")
		}

		iat, ok := claimAsInt64(raw)
		if !ok {
			return newError(ErrInvalidToken, "invalid iat claim")
		}

		if now > iat+maxAge+leeway {
			return newError(ErrExpired, "token has expired")
		}
	}

	return nil
}

func validateRegisteredClaims(claims map[string]any, opts VerifyOptions) error {
	for _, name := range opts.Required {
		if !hasClaim(claims, name) {
			return newError(ErrClaimMissing, "token is missing required claim")
		}
	}

	if opts.Issuer != "" {
		actual, ok := claimAsString(claims["iss"])
		if !ok || actual != opts.Issuer {
			return newError(ErrIssuerMismatch, "token issuer does not match")
		}
	}

	if opts.Audience != "" {
		if !audienceMatches(claims["aud"], opts.Audience) {
			return newError(ErrAudienceMismatch, "token audience does not match")
		}
	}

	if opts.Subject != "" {
		actual, ok := claimAsString(claims["sub"])
		if !ok || actual != opts.Subject {
			return newError(ErrSubjectMismatch, "token subject does not match")
		}
	}

	return nil
}
