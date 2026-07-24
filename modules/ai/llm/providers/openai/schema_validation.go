package openai

import (
	"unicode/utf8"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

const (
	maxStructuredOutputDepth         = 10
	maxStructuredOutputProperties    = 5000
	maxStructuredOutputStringLength  = 120000
	maxStructuredOutputEnumValues    = 1000
	largeStructuredOutputEnumSize    = 250
	maxLargeStructuredOutputEnumText = 15000
)

var unsupportedStructuredOutputKeywords = []string{
	"oneOf",
	"allOf",
	"not",
	"dependentRequired",
	"dependentSchemas",
	"if",
	"then",
	"else",
}

type structuredOutputSchemaStats struct {
	properties  int
	stringChars int
	enumValues  int
}

func validateStructuredOutputSchema(document map[string]any) error {
	if document == nil {
		return invalidStructuredOutputSchema("schema must have an object root")
	}
	if _, found := document["anyOf"]; found {
		return invalidStructuredOutputSchema("root-level anyOf is not supported")
	}
	if document["type"] != "object" {
		return invalidStructuredOutputSchema("schema must have an object root")
	}

	stats := &structuredOutputSchemaStats{}
	if err := validateStructuredOutputNode(document, 0, stats); err != nil {
		return err
	}
	if stats.properties > maxStructuredOutputProperties {
		return invalidStructuredOutputSchema("schema exceeds the property limit")
	}
	if stats.stringChars > maxStructuredOutputStringLength {
		return invalidStructuredOutputSchema("schema exceeds the string length limit")
	}
	if stats.enumValues > maxStructuredOutputEnumValues {
		return invalidStructuredOutputSchema("schema exceeds the enum value limit")
	}

	return nil
}

func validateStructuredOutputNode(
	node map[string]any,
	objectDepth int,
	stats *structuredOutputSchemaStats,
) error {
	for _, keyword := range unsupportedStructuredOutputKeywords {
		if _, found := node[keyword]; found {
			return invalidStructuredOutputSchema(keyword + " is not supported")
		}
	}

	if isObjectSchema(node) {
		objectDepth++
		if objectDepth > maxStructuredOutputDepth {
			return invalidStructuredOutputSchema("schema exceeds the nesting depth limit")
		}
		if additional, found := node["additionalProperties"]; !found || additional != false {
			return invalidStructuredOutputSchema("every object must set additionalProperties to false")
		}

		properties, err := schemaMap(node, "properties")
		if err != nil {
			return err
		}
		required, err := requiredPropertyNames(node)
		if err != nil {
			return err
		}
		for name := range properties {
			if _, found := required[name]; !found {
				return invalidStructuredOutputSchema("every property must be required")
			}
			stats.stringChars += utf8.RuneCountInString(name)
		}
		stats.properties += len(properties)
	}

	if enum, found := node["enum"].([]any); found {
		stats.enumValues += len(enum)
		enumChars := stringValueChars(enum)
		stats.stringChars += enumChars
		if len(enum) > largeStructuredOutputEnumSize && enumChars > maxLargeStructuredOutputEnumText {
			return invalidStructuredOutputSchema("schema exceeds the large enum string length limit")
		}
	}
	if constant, found := node["const"].(string); found {
		stats.stringChars += utf8.RuneCountInString(constant)
	}

	for _, keyword := range []string{"$defs", "definitions"} {
		definitions, err := schemaMap(node, keyword)
		if err != nil {
			return err
		}
		for name, definition := range definitions {
			stats.stringChars += utf8.RuneCountInString(name)
			if err := validateStructuredOutputChild(definition, objectDepth, stats); err != nil {
				return err
			}
		}
	}
	patternProperties, err := schemaMap(node, "patternProperties")
	if err != nil {
		return err
	}
	for _, property := range patternProperties {
		if err := validateStructuredOutputChild(property, objectDepth, stats); err != nil {
			return err
		}
	}

	properties, err := schemaMap(node, "properties")
	if err != nil {
		return err
	}
	for _, property := range properties {
		if err := validateStructuredOutputChild(property, objectDepth, stats); err != nil {
			return err
		}
	}

	for _, keyword := range []string{
		"additionalProperties",
		"contains",
		"contentSchema",
		"items",
		"propertyNames",
		"unevaluatedItems",
		"unevaluatedProperties",
	} {
		if child, found := node[keyword]; found {
			if _, booleanSchema := child.(bool); booleanSchema {
				continue
			}
			if err := validateStructuredOutputChild(child, objectDepth, stats); err != nil {
				return err
			}
		}
	}
	for _, keyword := range []string{"anyOf", "prefixItems"} {
		if children, found := node[keyword]; found {
			items, ok := children.([]any)
			if !ok {
				return invalidStructuredOutputSchema(keyword + " must be an array")
			}
			for _, child := range items {
				if err := validateStructuredOutputChild(child, objectDepth, stats); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateStructuredOutputChild(
	value any,
	objectDepth int,
	stats *structuredOutputSchemaStats,
) error {
	child, ok := value.(map[string]any)
	if !ok {
		return invalidStructuredOutputSchema("schema contains an invalid nested definition")
	}

	return validateStructuredOutputNode(child, objectDepth, stats)
}

func schemaMap(node map[string]any, keyword string) (map[string]any, error) {
	value, found := node[keyword]
	if !found {
		return nil, nil
	}

	result, ok := value.(map[string]any)
	if !ok {
		return nil, invalidStructuredOutputSchema(keyword + " must be an object")
	}

	return result, nil
}

func requiredPropertyNames(node map[string]any) (map[string]struct{}, error) {
	value, found := node["required"]
	if !found {
		return nil, nil
	}

	items, ok := value.([]any)
	if !ok {
		return nil, invalidStructuredOutputSchema("required must be an array")
	}

	result := make(map[string]struct{}, len(items))
	for _, item := range items {
		name, ok := item.(string)
		if !ok {
			return nil, invalidStructuredOutputSchema("required must contain property names")
		}
		result[name] = struct{}{}
	}

	return result, nil
}

func isObjectSchema(node map[string]any) bool {
	typeValue, found := node["type"]
	if !found {
		_, found = node["properties"]

		return found
	}
	if typeValue == "object" {
		return true
	}

	types, ok := typeValue.([]any)
	if !ok {
		return false
	}
	for _, value := range types {
		if value == "object" {
			return true
		}
	}

	return false
}

func stringValueChars(values []any) int {
	total := 0
	for _, value := range values {
		if text, ok := value.(string); ok {
			total += utf8.RuneCountInString(text)
		}
	}

	return total
}

func invalidStructuredOutputSchema(message string) error {
	return core.NewError(core.ErrInvalidSchema, message)
}
