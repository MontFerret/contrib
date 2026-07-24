package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

type testProviderFactory struct {
	name    string
	options core.ModelOptions
}

func (f *testProviderFactory) Name() string {
	return f.name
}

func (f *testProviderFactory) NewModel(_ context.Context, options core.ModelOptions) (core.Model, error) {
	f.options = options
	backend := newTestBackend()

	return core.NewStatelessModel(f.name, options.Model, backend, backend), nil
}
