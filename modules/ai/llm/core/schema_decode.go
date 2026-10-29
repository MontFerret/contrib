package core

import (
	"bytes"
	"context"
	"encoding/json"

	ferretjson "github.com/MontFerret/ferret/v2/pkg/encoding/json"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// DecodeSchema converts a Ferret object into a compiled JSON Schema.
func DecodeSchema(ctx context.Context, value runtime.Value) (Schema, error) {
	if _, ok := value.(runtime.Map); !ok {
		return Schema{}, NewError(ErrInvalidSchema, "schema must be an object")
	}

	var raw bytes.Buffer
	if err := ferretjson.Default.Encode(ctx, &raw, value); err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema is not valid JSON")
	}

	var document map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw.Bytes()))
	decoder.UseNumber()

	if err := decoder.Decode(&document); err != nil || decoderHasTrailingValue(decoder) {
		return Schema{}, NewError(ErrInvalidSchema, "schema is not valid JSON")
	}

	return NewSchema(document)
}
