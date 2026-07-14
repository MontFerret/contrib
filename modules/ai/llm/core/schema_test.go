package core

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestSchemaAllowsLocalReferencesAndPreservesFerretNumbers(t *testing.T) {
	schema, err := NewSchema(map[string]any{
		"type":  "object",
		"$defs": map[string]any{"count": map[string]any{"type": "integer"}},
		"properties": map[string]any{
			"count": map[string]any{"$ref": "#/$defs/count"},
		},
		"required": []string{"count"},
	})
	if err != nil {
		t.Fatal(err)
	}

	value, err := schema.ValidateJSON([]byte("{\"count\":42}"))
	if err != nil {
		t.Fatal(err)
	}
	countValue := objectValue(t, value, "count")
	if count, ok := countValue.(runtime.Int); !ok || count != 42 {
		t.Fatalf("expected Ferret Int, got %T %v", countValue, countValue)
	}
}

func TestSchemaRejectsExternalReferencesBeforeValidation(t *testing.T) {
	_, err := NewSchema(map[string]any{"$ref": "https://example.com/schema.json"})
	requireCode(t, err, ErrInvalidSchema)
	_, err = NewSchema(map[string]any{
		"properties": map[string]any{"x": map[string]any{"$ref": "file:///tmp/schema.json"}},
	})
	requireCode(t, err, ErrInvalidSchema)
}

func TestStructuredOutputErrorsAreDistinct(t *testing.T) {
	schema, err := NewSchema(map[string]any{"type": "object", "required": []string{"name"}})
	if err != nil {
		t.Fatal(err)
	}

	_, err = schema.ValidateJSON([]byte("not-json"))
	requireCode(t, err, ErrInvalidStructuredOutput)
	_, err = schema.ValidateJSON([]byte("{}"))
	requireCode(t, err, ErrSchemaValidation)
}

func TestDecodeSchemaRequiresObject(t *testing.T) {
	_, err := DecodeSchema(context.Background(), runtime.NewString("{}"))
	requireCode(t, err, ErrInvalidSchema)
}

func TestDecodeSchemaPreservesLargeIntegerConstraints(t *testing.T) {
	const exact = int64(9007199254740993)
	schema, err := DecodeSchema(context.Background(), object(map[string]runtime.Value{
		"type": runtime.NewString("object"),
		"properties": object(map[string]runtime.Value{
			"value": object(map[string]runtime.Value{
				"type":    runtime.NewString("integer"),
				"const":   runtime.NewInt64(exact),
				"minimum": runtime.NewInt64(exact),
				"maximum": runtime.NewInt64(exact),
			}),
		}),
		"required": runtime.NewArrayWith(runtime.NewString("value")),
	}))
	if err != nil {
		t.Fatal(err)
	}

	valueSchema := schema.Document()["properties"].(map[string]any)["value"].(map[string]any)
	for _, keyword := range []string{"const", "minimum", "maximum"} {
		number, ok := valueSchema[keyword].(json.Number)
		if !ok || number.String() != "9007199254740993" {
			t.Fatalf("%s lost integer precision: %T %v", keyword, valueSchema[keyword], valueSchema[keyword])
		}
	}

	if _, err := schema.ValidateJSON([]byte(`{"value":9007199254740993}`)); err != nil {
		t.Fatalf("validate exact integer: %v", err)
	}
	_, err = schema.ValidateJSON([]byte(`{"value":9007199254740992}`))
	requireCode(t, err, ErrSchemaValidation)
}

func TestSchemaDocumentReturnsDefensiveCopy(t *testing.T) {
	original := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}
	schema, err := NewSchema(original)
	if err != nil {
		t.Fatal(err)
	}

	original["type"] = "string"
	document := schema.Document()
	document["type"] = "array"
	document["properties"].(map[string]any)["name"].(map[string]any)["type"] = "integer"
	document["required"].([]any)[0] = "other"

	second := schema.Document()
	if second["type"] != "object" {
		t.Fatalf("schema root was mutated: %#v", second)
	}
	name := second["properties"].(map[string]any)["name"].(map[string]any)
	if name["type"] != "string" || second["required"].([]any)[0] != "name" {
		t.Fatalf("nested schema state was mutated: %#v", second)
	}
	if _, err := schema.ValidateJSON([]byte(`{"name":"Ada"}`)); err != nil {
		t.Fatalf("compiled validator diverged from document: %v", err)
	}
}
