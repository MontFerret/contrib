package core

import (
	"strings"
	"testing"
)

func TestChatBuilderAppendsQueryAsFinalUserMessage(t *testing.T) {
	request, err := BuildGenerationRequest(OperationRequest{
		Mode:  ModeChat,
		Input: "current",
		Semantic: SemanticOptions{Messages: []Message{
			TextMessage(RoleSystem, "system"),
			TextMessage(RoleAssistant, "prior"),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(request.Messages) != 3 {
		t.Fatalf("expected three messages, got %#v", request.Messages)
	}
	last := request.Messages[len(request.Messages)-1]
	if last.Role != RoleUser || last.Content.Text != "current" {
		t.Fatalf("query was not appended as final user message: %#v", last)
	}
}

func TestOperationInstructionsHaveDeterministicOrder(t *testing.T) {
	summary, err := BuildGenerationRequest(OperationRequest{
		Mode:  ModeSummarize,
		Input: "text",
		Semantic: SemanticOptions{
			Style:        "brief",
			MaxWords:     10,
			Instructions: "Caller instruction.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertOrdered(t, summary.Instructions,
		"Summarize the provided text.",
		"Use this summary style: brief.",
		"Use no more than 10 words.",
		"Caller instruction.",
	)

	classification, err := BuildStructuredRequest(OperationRequest{
		Mode: ModeClassify,
		Semantic: SemanticOptions{
			Labels:       []string{"yes", "no"},
			Instructions: "Caller instruction.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if classification.Name != "classification_result" {
		t.Fatalf("unexpected format name %q", classification.Name)
	}
	assertOrdered(t, classification.Instructions,
		"Classify the provided text",
		"Allowed labels:",
		"Caller instruction.",
	)
}

func TestExtractUsesStrictStableSchemaName(t *testing.T) {
	schema, err := NewSchema(map[string]any{"type": "object"})
	if err != nil {
		t.Fatal(err)
	}
	request, err := BuildStructuredRequest(OperationRequest{
		Mode: ModeExtract,
		Semantic: SemanticOptions{
			Schema:       schema,
			Instructions: "Caller instruction.",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Name != "extract_result" {
		t.Fatalf("unexpected format name %q", request.Name)
	}
	assertOrdered(t, request.Instructions,
		"Extract structured data from the provided text.",
		"Return only data that conforms to the supplied schema.",
		"Caller instruction.",
	)
}

func assertOrdered(t *testing.T, text string, fragments ...string) {
	t.Helper()

	offset := -1
	for _, fragment := range fragments {
		next := strings.Index(text, fragment)
		if next <= offset {
			t.Fatalf("%q did not occur in order in %q", fragment, text)
		}
		offset = next
	}
}
