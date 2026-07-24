package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
)

func TestExecutorSendsSupportedNestedSchemaUnchanged(t *testing.T) {
	t.Parallel()

	schema := compileSchema(t, map[string]any{
		"type": "object",
		"$defs": map[string]any{
			"address": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"city": map[string]any{"type": "string"},
				},
				"required":             []any{"city"},
				"additionalProperties": false,
			},
		},
		"properties": map[string]any{
			"address": map[string]any{"$ref": "#/$defs/address"},
			"nickname": map[string]any{
				"anyOf": []any{
					map[string]any{"type": "string"},
					map[string]any{"type": "null"},
				},
			},
		},
		"required":             []any{"address", "nickname"},
		"additionalProperties": false,
	})
	expected, err := json.Marshal(schema.Document())
	if err != nil {
		t.Fatal(err)
	}
	fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
		return jsonResponse(http.StatusOK, successResponseBody(`{"address":{"city":"London"},"nickname":null}`)), nil
	})
	model := newTestModel(t, "gpt-test")

	_, err = model.GenerateStructured(networkContext(t, fake), structuredRequest(schema))
	if err != nil {
		t.Fatalf("generate structured: %v", err)
	}
	payload := decodeRequestBody(t, fake.Requests()[0])
	actualSchema := payload["text"].(map[string]any)["format"].(map[string]any)["schema"]
	actual, err := json.Marshal(actualSchema)
	if err != nil {
		t.Fatal(err)
	}
	if string(actual) != string(expected) {
		t.Fatalf("provider schema changed:\nexpected: %s\nactual:   %s", expected, actual)
	}
}

func TestExecutorRejectsUnsupportedStructuredOutputSchemasBeforeHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		document map[string]any
		name     string
	}{
		{name: "non-object root", document: map[string]any{"type": "string"}},
		{name: "root anyOf", document: map[string]any{
			"type": "object", "anyOf": []any{map[string]any{"type": "object"}},
			"additionalProperties": false,
		}},
		{name: "missing additionalProperties", document: map[string]any{
			"type": "object", "properties": map[string]any{},
		}},
		{name: "property not required", document: map[string]any{
			"type": "object", "properties": map[string]any{"name": map[string]any{"type": "string"}},
			"additionalProperties": false,
		}},
		{name: "nested object missing additionalProperties", document: map[string]any{
			"type": "object", "properties": map[string]any{
				"nested": map[string]any{"type": "object", "properties": map[string]any{}},
			},
			"required": []any{"nested"}, "additionalProperties": false,
		}},
		{name: "property limit", document: schemaWithProperties(maxStructuredOutputProperties + 1)},
		{name: "nesting depth limit", document: schemaWithObjectDepth(maxStructuredOutputDepth + 1)},
		{name: "total string limit", document: map[string]any{
			"type": "object",
			"properties": map[string]any{
				strings.Repeat("x", maxStructuredOutputStringLength+1): map[string]any{"type": "string"},
			},
			"required": []any{strings.Repeat("x", maxStructuredOutputStringLength+1)}, "additionalProperties": false,
		}},
		{name: "enum value limit", document: schemaWithEnum(numericEnum(maxStructuredOutputEnumValues + 1))},
		{name: "large enum string limit", document: schemaWithEnum(stringEnum(
			largeStructuredOutputEnumSize+1,
			maxLargeStructuredOutputEnumText/(largeStructuredOutputEnumSize+1)+1,
		))},
	}

	for _, keyword := range unsupportedStructuredOutputKeywords {
		document := validStructuredOutputSchema()
		switch keyword {
		case "oneOf", "allOf":
			document[keyword] = []any{map[string]any{"type": "string"}}
		default:
			document[keyword] = map[string]any{}
		}
		tests = append(tests, struct {
			document map[string]any
			name     string
		}{name: "unsupported " + keyword, document: document})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			schema := compileSchema(t, test.document)
			fake := newFakeHTTPClient(func(context.Context, *ferrethttp.Request) (*ferrethttp.Response, error) {
				return nil, errors.New("unexpected OpenAI request")
			})
			model := newTestModel(t, "gpt-test")

			_, err := model.GenerateStructured(networkContext(t, fake), structuredRequest(schema))
			if code := errorCode(t, err); code != core.ErrInvalidSchema {
				t.Fatalf("expected %s, got %s", core.ErrInvalidSchema, code)
			}
			if len(fake.Requests()) != 0 {
				t.Fatalf("expected no provider request, got %d", len(fake.Requests()))
			}
		})
	}
}

func structuredRequest(schema core.Schema) core.StructuredRequest {
	return core.StructuredRequest{
		Request: core.Request{Messages: []core.Message{
			{Role: core.RoleUser, Content: core.Content{Type: core.ContentText, Text: "extract"}},
		}},
		Name:   "extract_result",
		Schema: schema,
	}
}

func compileSchema(t *testing.T, document map[string]any) core.Schema {
	t.Helper()

	schema, err := core.NewSchema(document)
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}

	return schema
}

func validStructuredOutputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           map[string]any{},
		"required":             []any{},
		"additionalProperties": false,
	}
}

func schemaWithProperties(count int) map[string]any {
	properties := make(map[string]any, count)
	required := make([]any, 0, count)
	for idx := range count {
		name := fmt.Sprintf("property_%d", idx)
		properties[name] = map[string]any{"type": "string"}
		required = append(required, name)
	}

	return map[string]any{
		"type": "object", "properties": properties, "required": required, "additionalProperties": false,
	}
}

func schemaWithObjectDepth(depth int) map[string]any {
	document := validStructuredOutputSchema()
	for idx := 1; idx < depth; idx++ {
		document = map[string]any{
			"type":                 "object",
			"properties":           map[string]any{"nested": document},
			"required":             []any{"nested"},
			"additionalProperties": false,
		}
	}

	return document
}

func schemaWithEnum(values []any) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"value": map[string]any{"enum": values},
		},
		"required":             []any{"value"},
		"additionalProperties": false,
	}
}

func numericEnum(count int) []any {
	values := make([]any, count)
	for idx := range count {
		values[idx] = idx
	}

	return values
}

func stringEnum(count, length int) []any {
	values := make([]any, count)
	for idx := range count {
		values[idx] = fmt.Sprintf("%06d%s", idx, strings.Repeat("x", length))
	}

	return values
}
