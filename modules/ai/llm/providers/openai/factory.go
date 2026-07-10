package openai

import (
	"context"
	"strings"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

// Factory creates OpenAI-backed stateless models.
type Factory struct{}

// NewFactory creates an OpenAI provider factory.
func NewFactory() *Factory {
	return &Factory{}
}

func (*Factory) Name() string {
	return "openai"
}

func (*Factory) NewModel(_ context.Context, options core.ModelOptions) (core.Model, error) {
	if strings.TrimSpace(options.Model) == "" {
		return nil, core.NewError(core.ErrInvalidOptions, "model must not be blank")
	}
	if strings.TrimSpace(options.APIKey) == "" {
		return nil, core.NewError(core.ErrInvalidOptions, "apiKey must not be blank")
	}

	executor := newExecutor(options.Model, options.APIKey)

	return core.NewStatelessModel("openai", options.Model, executor, executor), nil
}
