package openai

import (
	"context"
	"testing"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

func TestFactoryRequiresExplicitCredentialsAndOpaqueModel(t *testing.T) {
	t.Parallel()

	factory := NewFactory()
	if factory.Name() != "openai" {
		t.Fatalf("expected provider name openai, got %q", factory.Name())
	}

	tests := []struct {
		name    string
		options core.ModelOptions
	}{
		{name: "model", options: core.ModelOptions{APIKey: "key"}},
		{name: "api key", options: core.ModelOptions{Model: "gpt-test"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := factory.NewModel(context.Background(), test.options)
			if code := errorCode(t, err); code != core.ErrInvalidOptions {
				t.Fatalf("expected invalid options, got %s", code)
			}
		})
	}

	const modelName = "  vendor/model:preview  "
	model, err := factory.NewModel(context.Background(), core.ModelOptions{
		Model:  modelName,
		APIKey: "key",
	})
	if err != nil {
		t.Fatalf("create model: %v", err)
	}
	if model.ModelName() != modelName {
		t.Fatalf("expected opaque model name %q, got %q", modelName, model.ModelName())
	}
}
