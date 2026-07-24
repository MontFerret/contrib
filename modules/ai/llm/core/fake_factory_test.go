package core

import "context"

type fakeFactory struct {
	name     string
	executor *fakeExecutor
	options  ModelOptions
}

func (f *fakeFactory) Name() string {
	return f.name
}

func (f *fakeFactory) NewModel(_ context.Context, options ModelOptions) (Model, error) {
	f.options = options

	return NewStatelessModel(f.name, options.Model, f.executor, f.executor), nil
}
