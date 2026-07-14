package core

import (
	"bytes"
	"encoding/json"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"

	ferretjson "github.com/MontFerret/ferret/v2/pkg/encoding/json"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Schema is a locally compiled JSON Schema document suitable for a provider.
type Schema struct {
	compiled *jsonschema.Schema
	raw      json.RawMessage
}

// NewSchema copies and compiles a provider-neutral JSON Schema.
func NewSchema(document map[string]any) (Schema, error) {
	raw, err := json.Marshal(document)
	if err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema is not valid JSON")
	}

	var copied map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&copied); err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema is not valid JSON")
	}

	if hasExternalReference(copied) {
		return Schema{}, NewError(ErrInvalidSchema, "external schema references are not supported")
	}

	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	const location = "urn:ai-llm:schema"
	if err := compiler.AddResource(location, copied); err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema could not be compiled")
	}

	compiled, err := compiler.Compile(location)
	if err != nil {
		return Schema{}, NewError(ErrInvalidSchema, "schema could not be compiled")
	}

	return Schema{
		raw:      append(json.RawMessage(nil), raw...),
		compiled: compiled,
	}, nil
}

// Document returns a defensive copy of the provider-neutral JSON Schema.
func (s Schema) Document() map[string]any {
	if len(s.raw) == 0 {
		return nil
	}

	var document map[string]any
	decoder := json.NewDecoder(bytes.NewReader(s.raw))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil
	}

	return document
}

// ValidateJSON parses, validates, and converts structured output to a Ferret value.
func (s Schema) ValidateJSON(data []byte) (runtime.Value, error) {
	if s.compiled == nil {
		return runtime.None, NewError(ErrInvalidSchema, "schema is not compiled")
	}

	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil || decoderHasTrailingValue(decoder) {
		return runtime.None, NewError(ErrInvalidStructuredOutput, "provider returned malformed JSON")
	}

	if err := s.compiled.Validate(decoded); err != nil {
		return runtime.None, NewError(ErrSchemaValidation, "provider output does not match the schema")
	}

	value, err := ferretjson.Default.Decode(data)
	if err != nil {
		return runtime.None, NewError(ErrInvalidStructuredOutput, "provider returned malformed JSON")
	}

	return value, nil
}
