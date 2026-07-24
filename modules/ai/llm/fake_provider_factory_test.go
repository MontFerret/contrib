package llm

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

type fakeProviderFactory struct {
	backend *fakeBackend
}

func newFakeProviderFactory() *fakeProviderFactory {
	return &fakeProviderFactory{backend: newFakeBackend()}
}

func newFakeProviderFactoryWithGeneration(generation string) *fakeProviderFactory {
	return &fakeProviderFactory{backend: newFakeBackendWithGeneration(generation)}
}

func (*fakeProviderFactory) Name() string {
	return "openai"
}

func (f *fakeProviderFactory) NewModel(
	_ context.Context,
	options core.ModelOptions,
) (core.Model, error) {
	return core.NewStatelessModel("openai", options.Model, f.backend, f.backend), nil
}
