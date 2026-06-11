package core

import (
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func buildInspectResult(parsed *parsedToken) (runtime.Value, error) {
	header, err := runtime.ValueOf(parsed.header)
	if err != nil {
		return nil, err
	}

	claims, err := runtime.ValueOf(parsed.claims)
	if err != nil {
		return nil, err
	}

	return runtime.NewObjectWith(map[string]runtime.Value{
		"header":   header,
		"claims":   claims,
		"raw":      buildRawObject(parsed),
		"verified": runtime.False,
	}), nil
}

func buildVerifyResult(parsed *parsedToken) (runtime.Value, error) {
	header, err := runtime.ValueOf(parsed.header)
	if err != nil {
		return nil, err
	}

	claims, err := runtime.ValueOf(parsed.claims)
	if err != nil {
		return nil, err
	}

	return runtime.NewObjectWith(map[string]runtime.Value{
		"header":   header,
		"claims":   claims,
		"verified": runtime.True,
	}), nil
}

func buildRawObject(parsed *parsedToken) runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"header":    runtime.NewString(parsed.rawHeader),
		"payload":   runtime.NewString(parsed.rawPayload),
		"signature": runtime.NewString(parsed.rawSignature),
	})
}
