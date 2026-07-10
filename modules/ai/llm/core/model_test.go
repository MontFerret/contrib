package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestStatelessModelIsOpaqueAndPointerPreserving(t *testing.T) {
	model := testModel(&fakeExecutor{})
	if model.String() != "<ai.llm.model>" || fmt.Sprintf("%#v", model) != "<ai.llm.model>" {
		t.Fatalf("unsafe model display: %s / %#v", model.String(), model)
	}
	if model.Copy() != model {
		t.Fatal("Copy must preserve the immutable model pointer")
	}

	raw, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var display string
	if err := json.Unmarshal(raw, &display); err != nil {
		t.Fatal(err)
	}
	if display != "<ai.llm.model>" {
		t.Fatalf("unsafe model JSON: %s", raw)
	}

	debug := model.(runtime.DebugInspectable).DebugInfo()
	if debug.Display != "<ai.llm.model>" || debug.TypeName != "ai.llm.model" {
		t.Fatalf("unsafe debug info: %#v", debug)
	}
}

func TestStatelessModelSupportsConcurrentRequests(t *testing.T) {
	executor := &fakeExecutor{}
	model := testModel(executor)

	const requests = 16
	var wait sync.WaitGroup
	wait.Add(requests)
	errs := make(chan error, requests)
	for range requests {
		go func() {
			defer wait.Done()
			_, err := Execute(context.Background(), model, OperationRequest{Mode: ModeGenerate, Input: "prompt"})
			errs <- err
		}()
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := len(executor.Requests()); got != requests {
		t.Fatalf("expected %d provider requests, got %d", requests, got)
	}
}

func TestModelQueryPlainIsSingletonAndOneIsScalar(t *testing.T) {
	ctx := context.Background()
	executor := &fakeExecutor{generateFn: func(context.Context, Request) (Response, error) {
		return Response{Text: "answer"}, nil
	}}
	model := testModel(executor)
	query := runtime.Query{Expression: runtime.NewString("prompt")}

	list, err := model.Query(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	length, _ := list.Length(ctx)
	if length != 1 {
		t.Fatalf("expected singleton query result, got %d", length)
	}

	value, err := model.QueryOne(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if value.String() != "answer" {
		t.Fatalf("unexpected scalar result: %v", value)
	}

	_, err = model.QueryCount(ctx, query)
	requireCode(t, err, ErrUnsupportedOperation)
	_, err = model.QueryExists(ctx, query)
	requireCode(t, err, ErrUnsupportedOperation)
	if requests := executor.Requests(); len(requests) != 2 {
		t.Fatalf("COUNT and EXISTS must not invoke the provider, got %d total requests", len(requests))
	}
}
