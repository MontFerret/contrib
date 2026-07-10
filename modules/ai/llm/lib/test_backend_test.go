package lib

import (
	"context"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

type testBackend struct{}

func newTestBackend() *testBackend {
	return &testBackend{}
}

func (*testBackend) Generate(_ context.Context, _ core.Request) (core.Response, error) {
	return core.Response{Text: "generated"}, nil
}

func (*testBackend) GenerateStructured(
	_ context.Context,
	request core.StructuredRequest,
) (core.Response, error) {
	if request.Name == "classification_result" {
		return core.Response{Text: `{"label":"one"}`}, nil
	}

	return core.Response{Text: `{"value":"ok"}`}, nil
}
