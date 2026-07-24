package lib

import (
	"context"
	"errors"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestModelResolvesRegistryFromContext(t *testing.T) {
	factory := &testProviderFactory{name: "custom"}
	registry := core.NewRegistry()
	if err := registry.Register(factory); err != nil {
		t.Fatalf("unexpected provider registration error: %v", err)
	}

	ctx := core.WithRegistry(context.Background(), registry)
	value, err := Model(
		ctx,
		runtime.NewString(" CUSTOM "),
		modelOptionsValue(),
	)
	if err != nil {
		t.Fatalf("unexpected model creation error: %v", err)
	}

	model, ok := value.(core.Model)
	if !ok {
		t.Fatalf("expected core.Model, got %T", value)
	}
	if model.Provider() != "custom" {
		t.Fatalf("expected custom provider, got %q", model.Provider())
	}
	if model.ModelName() != "opaque/model-name" {
		t.Fatalf("expected opaque model name, got %q", model.ModelName())
	}
	if factory.options.APIKey != "explicit-test-key" {
		t.Fatal("expected decoded options to reach the provider factory")
	}
}

func TestModelReturnsRegistryNotFound(t *testing.T) {
	value, err := Model(
		context.Background(),
		runtime.NewString("custom"),
		modelOptionsValue(),
	)
	if value != runtime.None {
		t.Fatalf("expected none value, got %v", value)
	}
	if !errors.Is(err, core.ErrRegistryNotFound) {
		t.Fatalf("expected ErrRegistryNotFound, got %v", err)
	}
}

func modelOptionsValue() runtime.Value {
	return runtime.NewObjectWith(map[string]runtime.Value{
		"model":  runtime.NewString("opaque/model-name"),
		"apiKey": runtime.NewString("explicit-test-key"),
	})
}
