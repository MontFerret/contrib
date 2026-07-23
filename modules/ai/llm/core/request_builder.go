package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BuildGenerationRequest deterministically builds a text provider request.
func BuildGenerationRequest(operation OperationRequest) (Request, error) {
	request := Request{Options: operation.Execution}

	switch operation.Mode {
	case ModeGenerate:
		request.Messages = []Message{TextMessage(RoleUser, operation.Input)}
		request.Instructions = operation.Semantic.Instructions
	case ModeChat:
		request.Messages = copyMessages(operation.Semantic.Messages)
		request.Messages = append(request.Messages, TextMessage(RoleUser, operation.Input))
		request.Instructions = operation.Semantic.Instructions
	case ModeSummarize:
		request.Messages = []Message{TextMessage(RoleUser, operation.Input)}
		constraints := make([]string, 0, 2)

		if operation.Semantic.Style != "" {
			constraints = append(constraints, "Use this summary style: "+operation.Semantic.Style+".")
		}

		if operation.Semantic.MaxWords > 0 {
			constraints = append(constraints, fmt.Sprintf("Use no more than %d words.", operation.Semantic.MaxWords))
		}

		request.Instructions = joinInstructions(
			"Summarize the provided text.",
			strings.Join(constraints, "\n"),
			operation.Semantic.Instructions,
		)
	default:
		return Request{}, NewError(ErrUnsupportedOperation, "operation does not produce text output")
	}

	return request, nil
}

// BuildStructuredRequest deterministically builds a structured provider request.
func BuildStructuredRequest(operation OperationRequest) (StructuredRequest, error) {
	request := StructuredRequest{
		Request: Request{
			Messages: []Message{TextMessage(RoleUser, operation.Input)},
			Options:  operation.Execution,
		},
	}

	switch operation.Mode {
	case ModeExtract:
		if operation.Semantic.Schema.compiled == nil {
			return StructuredRequest{}, NewError(ErrInvalidSchema, "extract schema is required")
		}

		request.Name = "extract_result"
		request.Description = "Structured data extracted from the provided text."
		request.Schema = operation.Semantic.Schema
		request.Instructions = joinInstructions(
			"Extract structured data from the provided text.",
			"Return only data that conforms to the supplied schema.",
			operation.Semantic.Instructions,
		)
	case ModeClassify:
		if err := validateLabels(operation.Semantic.Labels); err != nil {
			return StructuredRequest{}, err
		}

		document := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"label": map[string]any{
					"type": "string",
					"enum": append([]string(nil), operation.Semantic.Labels...),
				},
			},
			"required":             []string{"label"},
			"additionalProperties": false,
		}
		schema, err := NewSchema(document)

		if err != nil {
			return StructuredRequest{}, err
		}

		encodedLabels, _ := json.Marshal(operation.Semantic.Labels)
		request.Name = "classification_result"
		request.Description = "One selected classification label."
		request.Schema = schema
		request.Instructions = joinInstructions(
			"Classify the provided text using exactly one allowed label.",
			"Allowed labels: "+string(encodedLabels)+".",
			operation.Semantic.Instructions,
		)
	default:
		return StructuredRequest{}, NewError(ErrUnsupportedOperation, "operation does not produce structured output")
	}

	return request, nil
}
