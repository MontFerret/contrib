package core

import (
	"context"
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
