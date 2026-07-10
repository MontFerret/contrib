package llm

import "github.com/MontFerret/contrib/modules/ai/llm/core"

type options struct {
	providerFactories []core.ProviderFactory
}

// Option configures the AI::LLM module during engine construction.
type Option func(*options)

// WithProviderFactory adds a provider factory to the module. A factory with
// the same provider name as a built-in provider deliberately replaces that
// built-in, which is useful for embedding and deterministic tests.
func WithProviderFactory(factory core.ProviderFactory) Option {
	return func(opts *options) {
		opts.providerFactories = append(opts.providerFactories, factory)
	}
}
