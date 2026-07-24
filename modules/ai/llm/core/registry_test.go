package core

import (
	"context"
	"testing"
)

func TestRegistryNormalizesProviderAndReplacesExplicitly(t *testing.T) {
	registry := NewRegistry()
	first := &fakeFactory{name: " OpenAI ", executor: &fakeExecutor{}}
	if err := registry.Register(first); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(first); err == nil {
		t.Fatal("expected duplicate registration error")
	}

	replacement := &fakeFactory{name: "openai", executor: &fakeExecutor{}}
	if err := registry.Replace(replacement); err != nil {
		t.Fatal(err)
	}

	model, err := registry.NewModel(context.Background(), " OPENAI ", ModelOptions{Model: "gpt-opaque", APIKey: "key"})
	if err != nil {
		t.Fatal(err)
	}
	if model.ModelName() != "gpt-opaque" || replacement.options.Model != "gpt-opaque" {
		t.Fatalf("model was not passed through unchanged: %#v", replacement.options)
	}
}

func TestRegistryRejectsUnsupportedProvider(t *testing.T) {
	_, err := NewRegistry().NewModel(context.Background(), "missing", ModelOptions{})
	requireCode(t, err, ErrUnsupportedProvider)
}

func TestRegistryRejectsNilFactories(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(nil); err == nil {
		t.Fatal("expected nil factory rejection")
	}

	var factory *fakeFactory
	if err := registry.Register(factory); err == nil {
		t.Fatal("expected typed nil factory rejection")
	}
}
