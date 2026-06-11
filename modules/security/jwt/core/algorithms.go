package core

import (
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

func signingMethod(alg string) (jwt.SigningMethod, error) {
	method := jwt.GetSigningMethod(alg)
	if method == nil {
		return nil, newError(ErrUnsupportedAlgorithm, "unsupported JWT algorithm")
	}

	if alg == "none" {
		return nil, newError(ErrUnexpectedAlgorithm, "alg none is not allowed")
	}

	return method, nil
}

func verifyKeyMaterial(key any, alg string) (any, error) {
	if err := verifyKeyMatchesAlgorithm(key, alg); err != nil {
		return nil, err
	}

	switch typed := key.(type) {
	case *HMACKey:
		return typed.Secret(), nil
	case *PublicKey:
		switch typed.family {
		case familyRSA:
			return typed.rsaKey, nil
		case familyECDSA:
			return typed.ecKey, nil
		case familyEd25519:
			return typed.edKey, nil
		}
	case *PrivateKey:
		switch typed.family {
		case familyRSA:
			return typed.rsaKey, nil
		case familyECDSA:
			return typed.ecKey, nil
		case familyEd25519:
			return typed.edKey, nil
		}
	}

	return nil, newError(ErrInvalidKey, "unsupported verification key")
}

func signKeyMaterial(key any, alg string) (any, error) {
	if err := verifyKeyMatchesAlgorithm(key, alg); err != nil {
		return nil, err
	}

	switch typed := key.(type) {
	case *HMACKey:
		return typed.Secret(), nil
	case *PrivateKey:
		switch typed.family {
		case familyRSA:
			return typed.rsaKey, nil
		case familyECDSA:
			return typed.ecKey, nil
		case familyEd25519:
			return typed.edKey, nil
		}
	}

	return nil, newError(ErrInvalidKey, "signing requires an HMAC or private key")
}

func normalizeAlgorithms(algorithms []string) []string {
	out := make([]string, 0, len(algorithms))
	for _, alg := range algorithms {
		trimmed := strings.TrimSpace(alg)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}
