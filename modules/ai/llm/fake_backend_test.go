package llm

import (
	"context"
	"sync"

	"github.com/MontFerret/contrib/modules/ai/llm/core"
)

type fakeBackend struct {
	generation         string
	requests           []core.Request
	structuredRequests []core.StructuredRequest
	mu                 sync.Mutex
}

func newFakeBackend() *fakeBackend {
	return &fakeBackend{}
}

func newFakeBackendWithGeneration(generation string) *fakeBackend {
	return &fakeBackend{generation: generation}
}

func (f *fakeBackend) Generate(_ context.Context, request core.Request) (core.Response, error) {
	f.mu.Lock()
	f.requests = append(f.requests, request)
	f.mu.Unlock()

	text := f.generation
	if text == "" {
		text = "generated response"
	}
	if len(request.Messages) > 0 {
		switch request.Messages[len(request.Messages)-1].Content.Text {
		case "first turn":
			text = "first answer"
		case "second turn":
			text = "second answer"
		case "query scalar":
			text = "scalar answer"
		}
	}

	return core.Response{ID: "response-id", Model: "opaque-model", Text: text}, nil
}

func (f *fakeBackend) GenerateStructured(
	_ context.Context,
	request core.StructuredRequest,
) (core.Response, error) {
	f.mu.Lock()
	f.structuredRequests = append(f.structuredRequests, request)
	f.mu.Unlock()

	text := `{"name":"Ada","score":9}`
	if request.Name == "classification_result" {
		text = `{"label":"billing"}`
	}

	return core.Response{ID: "structured-id", Model: "opaque-model", Text: text}, nil
}
