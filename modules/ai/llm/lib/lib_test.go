package lib

import (
	"slices"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestRegisterLib(t *testing.T) {
	library := runtime.NewLibrary()
	RegisterLib(library.Namespace("AI").Namespace("LLM"), core.NewRegistry())

	functions, err := library.Build()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}

	expected := []string{
		"AI::LLM::CHAT",
		"AI::LLM::CLASSIFY",
		"AI::LLM::EXTRACT",
		"AI::LLM::FORK",
		"AI::LLM::GENERATE",
		"AI::LLM::HISTORY",
		"AI::LLM::MODEL",
		"AI::LLM::RESET",
		"AI::LLM::SESSION",
		"AI::LLM::SUMMARIZE",
	}

	if functions.Size() != len(expected) {
		t.Fatalf("expected %d registered functions, got %d", len(expected), functions.Size())
	}

	names := functions.List()
	slices.Sort(names)
	slices.Sort(expected)
	if !slices.Equal(names, expected) {
		t.Fatalf("unexpected registered names: got %v, want %v", names, expected)
	}
}
