package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestOperationFunctionsValidateArity(t *testing.T) {
	target := testTarget()
	empty := runtime.NewObject()

	tests := []struct {
		name     string
		function runtime.Function
		valid    []runtime.Value
	}{
		{name: "GENERATE", function: Generate, valid: []runtime.Value{target, runtime.NewString("input")}},
		{name: "CHAT", function: Chat, valid: []runtime.Value{target, runtime.NewString("input")}},
		{name: "SUMMARIZE", function: Summarize, valid: []runtime.Value{target, runtime.NewString("input")}},
		{name: "EXTRACT", function: Extract, valid: []runtime.Value{target, runtime.NewString("input"), empty}},
		{name: "CLASSIFY", function: Classify, valid: []runtime.Value{target, runtime.NewString("input"), runtime.NewArrayWith(runtime.NewString("one"))}},
	}

	for _, test := range tests {
		t.Run(test.name+" missing", func(t *testing.T) {
			if _, err := test.function(context.Background()); err == nil {
				t.Fatal("expected missing-argument error")
			}
		})

		t.Run(test.name+" extra", func(t *testing.T) {
			args := append(append([]runtime.Value(nil), test.valid...), empty, empty)
			if _, err := test.function(context.Background(), args...); err == nil {
				t.Fatal("expected extra-argument error")
			}
		})
	}
}

func TestOperationFunctionsValidateRuntimeValues(t *testing.T) {
	target := testTarget()

	if _, err := Generate(context.Background(), runtime.NewString("not a model"), runtime.NewString("input")); err == nil {
		t.Fatal("expected target conversion error")
	}
	if _, err := Generate(context.Background(), target, runtime.NewInt(1)); err == nil {
		t.Fatal("expected input conversion error")
	} else if !errors.Is(err, runtime.ErrInvalidType) {
		t.Fatalf("expected SDK argument type error, got %v", err)
	}
	if _, err := Generate(context.Background(), target, runtime.None); err == nil {
		t.Fatal("expected none input conversion error")
	} else if !errors.Is(err, runtime.ErrInvalidType) {
		t.Fatalf("expected SDK argument type error for none, got %v", err)
	}
	if _, err := Extract(context.Background(), target, runtime.NewString("input"), runtime.NewString("not a schema")); err == nil {
		t.Fatal("expected schema conversion error")
	}
	if _, err := Classify(context.Background(), target, runtime.NewString("input"), runtime.NewString("not labels")); err == nil {
		t.Fatal("expected labels conversion error")
	}
	if _, err := Generate(
		context.Background(),
		target,
		runtime.NewString("input"),
		runtime.NewObjectWith(map[string]runtime.Value{"unknown": runtime.True}),
	); err == nil {
		t.Fatal("expected unknown function option error")
	} else if code, ok := core.CodeOf(err); !ok || code != core.ErrInvalidOptions {
		t.Fatalf("expected invalid-options error, got %v", err)
	}
}

func TestOperationFunctionsReturnFerretValues(t *testing.T) {
	target := testTarget()

	generated, err := Generate(context.Background(), target, runtime.NewString("input"))
	if err != nil {
		t.Fatalf("unexpected generation error: %v", err)
	}
	if generated != runtime.NewString("generated") {
		t.Fatalf("unexpected generation result: %v", generated)
	}

	summarized, err := Summarize(
		context.Background(),
		target,
		runtime.NewString("input"),
		runtime.NewObjectWith(map[string]runtime.Value{
			"style":           runtime.NewString("concise"),
			"maxWords":        runtime.NewInt(20),
			"instructions":    runtime.NewString("Prefer active voice."),
			"temperature":     runtime.NewFloat(0),
			"maxOutputTokens": runtime.NewInt(100),
			"timeout":         runtime.NewInt(0),
		}),
	)
	if err != nil {
		t.Fatalf("unexpected summarization error: %v", err)
	}
	if summarized != runtime.NewString("generated") {
		t.Fatalf("unexpected summarization result: %v", summarized)
	}

	schema := runtime.NewObjectWith(map[string]runtime.Value{
		"type": runtime.NewString("object"),
		"properties": runtime.NewObjectWith(map[string]runtime.Value{
			"value": runtime.NewObjectWith(map[string]runtime.Value{
				"type": runtime.NewString("string"),
			}),
		}),
		"required":             runtime.NewArrayWith(runtime.NewString("value")),
		"additionalProperties": runtime.False,
	})
	extracted, err := Extract(context.Background(), target, runtime.NewString("input"), schema)
	if err != nil {
		t.Fatalf("unexpected extraction error: %v", err)
	}
	object, err := runtime.CastObject(extracted)
	if err != nil {
		t.Fatalf("expected extraction object: %v", err)
	}
	value, err := object.Get(context.Background(), runtime.NewString("value"))
	if err != nil || value != runtime.NewString("ok") {
		t.Fatalf("unexpected extracted value: %v, %v", value, err)
	}

	classified, err := Classify(
		context.Background(),
		target,
		runtime.NewString("input"),
		runtime.NewArrayWith(runtime.NewString("one"), runtime.NewString("two")),
	)
	if err != nil {
		t.Fatalf("unexpected classification error: %v", err)
	}
	object, err = runtime.CastObject(classified)
	if err != nil {
		t.Fatalf("expected classification object: %v", err)
	}
	label, err := object.Get(context.Background(), runtime.NewString("label"))
	if err != nil || label != runtime.NewString("one") {
		t.Fatalf("unexpected classification label: %v, %v", label, err)
	}
}

func testTarget() core.Model {
	backend := newTestBackend()

	return core.NewStatelessModel("test", "test-model", backend, backend)
}
