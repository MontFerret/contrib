package core

import (
	"context"

	commonresource "github.com/MontFerret/contrib/pkg/common/resource"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

// StatelessModel is an immutable provider-neutral model handle.
type StatelessModel struct {
	generator  Generator
	structured StructuredGenerator
	provider   string
	model      string
	id         uint64
}

// NewStatelessModel creates an opaque model backed by provider executors.
func NewStatelessModel(provider, model string, generator Generator, structured StructuredGenerator) *StatelessModel {
	return &StatelessModel{
		provider:   provider,
		model:      model,
		generator:  generator,
		structured: structured,
		id:         newResourceID(),
	}
}

func (m *StatelessModel) Provider() string {
	return m.provider
}

func (m *StatelessModel) ModelName() string {
	return m.model
}

func (m *StatelessModel) Generate(ctx context.Context, request Request) (Response, error) {
	if m.generator == nil {
		return Response{}, NewError(ErrUnsupportedOperation, "text generation is not supported")
	}

	return m.generator.Generate(ctx, request)
}

func (m *StatelessModel) GenerateStructured(ctx context.Context, request StructuredRequest) (Response, error) {
	if m.structured == nil {
		return Response{}, NewError(ErrUnsupportedOperation, "structured generation is not supported")
	}

	return m.structured.GenerateStructured(ctx, request)
}

func (m *StatelessModel) Query(ctx context.Context, query runtime.Query) (runtime.List, error) {
	value, err := ExecuteQuery(ctx, m, query)
	if err != nil {
		return nil, err
	}

	return runtime.NewArrayWith(value), nil
}

func (m *StatelessModel) QueryOne(ctx context.Context, query runtime.Query) (runtime.Value, error) {
	return ExecuteQuery(ctx, m, query)
}

func (m *StatelessModel) QueryCount(context.Context, runtime.Query) (runtime.Int, error) {
	return runtime.ZeroInt, NewError(ErrUnsupportedOperation, "QUERY COUNT is not supported")
}

func (m *StatelessModel) QueryExists(context.Context, runtime.Query) (runtime.Boolean, error) {
	return runtime.False, NewError(ErrUnsupportedOperation, "QUERY EXISTS is not supported")
}

func (m *StatelessModel) ResourceID() uint64 {
	return m.id
}

func (m *StatelessModel) String() string {
	return commonresource.Display("ai.llm.model")
}

func (m *StatelessModel) GoString() string {
	return m.String()
}

func (m *StatelessModel) Hash() uint64 {
	return commonresource.Hash("ai.llm.model", m.id)
}

func (m *StatelessModel) Copy() runtime.Value {
	return m
}

func (m *StatelessModel) MarshalJSON() ([]byte, error) {
	return commonresource.MarshalDisplayJSON("ai.llm.model")
}

func (m *StatelessModel) DebugInfo() runtime.DebugInfo {
	return runtime.DebugInfo{TypeName: "ai.llm.model", Display: m.String()}
}
