package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// Model creates a provider model using the registry installed in the execution context.
func Model(ctx context.Context, providerValue, optionsValue runtime.Value) (runtime.Value, error) {
	provider, err := runtime.CastString(providerValue)
	if err != nil {
		return runtime.None, err
	}

	options, err := core.DecodeModelOptions(ctx, optionsValue)
	if err != nil {
		return runtime.None, err
	}

	registry, err := core.RegistryFrom(ctx)
	if err != nil {
		return runtime.None, err
	}

	model, err := registry.NewModel(ctx, provider.String(), options)
	if err != nil {
		return runtime.None, err
	}

	if !options.Session {
		return model, nil
	}

	return core.NewLocalSession(ctx, model, core.SessionOptions{})
}
