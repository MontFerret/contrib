package core

import (
	"context"
	"encoding/json"

	ferretjson "github.com/MontFerret/ferret/v2/pkg/encoding/json"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeSchema converts a Ferret object into a compiled JSON Schema.
func DecodeSchema(_ context.Context, value runtime.Value) (Schema, error) {
	if _, ok := value.(runtime.Map); !ok {
		return Schema{}, NewError(ErrInvalidSchema, "schema must be an object")
	}

	raw, err := ferretjson.Default.Encode(value)
	if err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema is not valid JSON")
	}

	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema is not valid JSON")
	}

	return NewSchema(document)
}
