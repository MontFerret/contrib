package core

import (
	"context"
	"sync"
)

type fakeExecutor struct {
	generateFn       func(context.Context, Request) (Response, error)
	generateStructFn func(context.Context, StructuredRequest) (Response, error)
	requests         []Request
	structured       []StructuredRequest
	mu               sync.Mutex
}

func (f *fakeExecutor) Generate(ctx context.Context, request Request) (Response, error) {
	f.mu.Lock()
	f.requests = append(f.requests, request)
	callback := f.generateFn
	f.mu.Unlock()

	if callback != nil {
		return callback(ctx, request)
	}

	return Response{Text: "ok"}, nil
}

func (f *fakeExecutor) GenerateStructured(ctx context.Context, request StructuredRequest) (Response, error) {
	f.mu.Lock()
	f.structured = append(f.structured, request)
	callback := f.generateStructFn
	f.mu.Unlock()

	if callback != nil {
		return callback(ctx, request)
	}

	return Response{Text: "{}"}, nil
}

func (f *fakeExecutor) Requests() []Request {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]Request(nil), f.requests...)
}

func (f *fakeExecutor) StructuredRequests() []StructuredRequest {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]StructuredRequest(nil), f.structured...)
}
