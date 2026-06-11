package core

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

type parsedToken struct {
	rawHeader    string
	rawPayload   string
	rawSignature string
	header       map[string]any
	claims       map[string]any
	compact      string
}

func parseCompactToken(token string, maxSize int) (*parsedToken, error) {
	if len(token) > maxSize {
		return nil, newError(ErrInvalidToken, "token exceeds maximum size")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, newError(ErrInvalidToken, "malformed compact JWT")
	}

	for _, part := range parts {
		if part == "" {
			return nil, newError(ErrInvalidToken, "malformed compact JWT")
		}
	}

	header, err := decodeSegment(parts[0])
	if err != nil {
		return nil, wrapError(ErrInvalidToken, "invalid JWT header", err)
	}

	claims, err := decodeSegment(parts[1])
	if err != nil {
		return nil, wrapError(ErrInvalidToken, "invalid JWT payload", err)
	}

	return &parsedToken{
		rawHeader:    parts[0],
		rawPayload:   parts[1],
		rawSignature: parts[2],
		header:       header,
		claims:       claims,
		compact:      token,
	}, nil
}

func decodeSegment(segment string) (map[string]any, error) {
	data, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	if out == nil {
		out = map[string]any{}
	}

	return out, nil
}

func headerAlgorithm(header map[string]any) (string, error) {
	raw, ok := header["alg"]
	if !ok {
		return "", newError(ErrInvalidToken, "JWT header missing alg")
	}

	alg, ok := raw.(string)
	if !ok || alg == "" {
		return "", newError(ErrInvalidToken, "JWT header has invalid alg")
	}

	return alg, nil
}

func algorithmAllowed(alg string, allowed []string) bool {
	for _, candidate := range allowed {
		if candidate == alg {
			return true
		}
	}

	return false
}
