package llm

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/contrib/modules/ai/llm/lib"
	openai "github.com/MontFerret/contrib/modules/ai/llm/providers/openai"
	"github.com/MontFerret/ferret/v2/pkg/module"
)

type mod struct {
	providerFactories []core.ProviderFactory
}

// New returns the AI::LLM module. Custom factories deliberately replace a
// built-in provider with the same normalized name.
func New(opts ...Option) module.Module {
	config := &options{}
	for _, apply := range opts {
		if apply != nil {
			apply(config)
		}
	}

	return &mod{providerFactories: append([]core.ProviderFactory(nil), config.providerFactories...)}
}

func (m *mod) Name() string {
	return "ai/llm"
}

func (m *mod) Register(bootstrap module.Bootstrap) error {
	registry := core.NewRegistry()
	if err := registry.Register(openai.NewFactory()); err != nil {
		return err
	}

	for _, factory := range m.providerFactories {
		if err := registry.Replace(factory); err != nil {
			return err
		}
	}

	lib.RegisterLib(bootstrap.Host().Library().Namespace("AI").Namespace("LLM"))
	bootstrap.Hooks().Session().BeforeRun(func(ctx context.Context) (context.Context, error) {
		ctx = core.WithRegistry(ctx, registry)

		return core.WithSessionScope(ctx), nil
	})
	bootstrap.Hooks().Session().AfterRun(func(ctx context.Context, _ error) error {
		return core.CloseSessionScope(ctx)
	})

	return nil
}
