package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestLocalSessionIsOpaqueAndPointerPreserving(t *testing.T) {
	session, err := NewLocalSession(context.Background(), testModel(&fakeExecutor{}), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if session.String() != "<ai.llm.session>" || fmt.Sprintf("%#v", session) != "<ai.llm.session>" {
		t.Fatalf("unsafe session display: %s / %#v", session.String(), session)
	}
	if session.Copy() != session {
		t.Fatal("Copy must preserve the session pointer")
	}

	raw, err := json.Marshal(session)
	if err != nil {
		t.Fatal(err)
	}
	var display string
	if err := json.Unmarshal(raw, &display); err != nil {
		t.Fatal(err)
	}
	if display != "<ai.llm.session>" {
		t.Fatalf("unsafe session JSON: %s", raw)
	}
	debug := session.(runtime.DebugInspectable).DebugInfo()
	if debug.TypeName != "ai.llm.session" || debug.Display != "<ai.llm.session>" {
		t.Fatalf("unsafe debug info: %#v", debug)
	}
}

func TestLocalSessionCommitsOnlyAfterFullSuccess(t *testing.T) {
	responses := []string{"not-json", "{\"name\":7}", "{\"name\":\"Ada\"}"}
	executor := &fakeExecutor{generateStructFn: func(context.Context, StructuredRequest) (Response, error) {
		text := responses[0]
		responses = responses[1:]

		return Response{Text: text}, nil
	}}
	session, err := NewLocalSession(context.Background(), testModel(executor), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := NewSchema(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []string{"name"},
	})
	if err != nil {
		t.Fatal(err)
	}
	operation := OperationRequest{Mode: ModeExtract, Input: "Ada", Semantic: SemanticOptions{Schema: schema}}

	_, err = Execute(context.Background(), session, operation)
	requireCode(t, err, ErrInvalidStructuredOutput)
	if history := session.History(); len(history) != 0 {
		t.Fatalf("failed output was committed: %#v", history)
	}

	_, err = Execute(context.Background(), session, operation)
	requireCode(t, err, ErrSchemaValidation)
	if history := session.History(); len(history) != 0 {
		t.Fatalf("schema-invalid output was committed: %#v", history)
	}

	value, err := Execute(context.Background(), session, operation)
	if err != nil {
		t.Fatal(err)
	}
	if objectValue(t, value, "name").String() != "Ada" {
		t.Fatalf("unexpected structured value: %v", value)
	}
	history := session.History()
	if len(history) != 2 || history[1].Content.Text != "{\"name\":\"Ada\"}" {
		t.Fatalf("structured JSON text was not committed: %#v", history)
	}
}

func TestLocalSessionProviderFailureIsTransactional(t *testing.T) {
	executor := &fakeExecutor{generateFn: func(context.Context, Request) (Response, error) {
		return Response{}, NewError(ErrRateLimit, "rate limited")
	}}
	session, err := NewLocalSession(context.Background(), testModel(executor), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = Execute(context.Background(), session, OperationRequest{Mode: ModeGenerate, Input: "prompt"})
	requireCode(t, err, ErrRateLimit)
	if len(session.History()) != 0 {
		t.Fatal("provider failure mutated history")
	}
}

func TestLocalSessionReplaysHistoryAndOrdersInstructions(t *testing.T) {
	executor := &fakeExecutor{generateFn: func(_ context.Context, request Request) (Response, error) {
		return Response{Text: "response"}, nil
	}}
	session, err := NewLocalSession(context.Background(), testModel(executor), SessionOptions{Instructions: "persistent"})
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range []string{"first", "second"} {
		_, err := Execute(context.Background(), session, OperationRequest{
			Mode:  ModeGenerate,
			Input: input,
			Semantic: SemanticOptions{
				Instructions: "per-operation",
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	requests := executor.Requests()
	if len(requests) != 2 || len(requests[1].Messages) != 3 {
		t.Fatalf("history was not replayed: %#v", requests)
	}
	if requests[1].Messages[0].Content.Text != "first" ||
		requests[1].Messages[1].Role != RoleAssistant ||
		requests[1].Messages[2].Content.Text != "second" {
		t.Fatalf("unexpected replay order: %#v", requests[1].Messages)
	}
	if requests[0].Instructions != "persistent\n\nper-operation" {
		t.Fatalf("unexpected instruction order: %q", requests[0].Instructions)
	}
}

func TestLocalSessionResetForkAndHistoryCopies(t *testing.T) {
	executor := &fakeExecutor{}
	ctx := WithSessionScope(context.Background())
	session, err := NewLocalSession(ctx, testModel(executor), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Execute(ctx, session, OperationRequest{Mode: ModeGenerate, Input: "first"}); err != nil {
		t.Fatal(err)
	}

	fork, err := session.Fork(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if fork.ResourceID() == session.ResourceID() {
		t.Fatal("fork must have an independent resource ID")
	}
	if _, err := Execute(ctx, fork, OperationRequest{Mode: ModeGenerate, Input: "fork-only"}); err != nil {
		t.Fatal(err)
	}
	if len(session.History()) != 2 || len(fork.History()) != 4 {
		t.Fatalf("fork histories are not independent: original=%#v fork=%#v", session.History(), fork.History())
	}

	copy := session.History()
	copy[0].Content.Text = "mutated"
	if session.History()[0].Content.Text != "first" {
		t.Fatal("History returned aliased state")
	}
	if err := session.Reset(); err != nil {
		t.Fatal(err)
	}
	if len(session.History()) != 0 || len(fork.History()) != 4 {
		t.Fatal("reset leaked across fork")
	}

	scope, ok := SessionScopeFrom(ctx)
	if !ok || scope.Len() != 2 {
		t.Fatalf("scope did not track session and fork: %v, %d", ok, scope.Len())
	}
}

func TestSessionScopeClosesAndClearsTrackedSessions(t *testing.T) {
	ctx := WithSessionScope(context.Background())
	session, err := NewLocalSession(ctx, testModel(&fakeExecutor{}), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.Fork(ctx); err != nil {
		t.Fatal(err)
	}
	scope, _ := SessionScopeFrom(ctx)
	if scope.Len() != 2 {
		t.Fatalf("expected two tracked sessions, got %d", scope.Len())
	}

	if err := CloseSessionScope(ctx); err != nil {
		t.Fatal(err)
	}
	if scope.Len() != 0 {
		t.Fatalf("scope was not cleared: %d", scope.Len())
	}
	if err := CloseSessionScope(ctx); err != nil {
		t.Fatalf("scope close must be idempotent: %v", err)
	}
	_, err = Execute(ctx, session, OperationRequest{Mode: ModeGenerate, Input: "closed"})
	requireCode(t, err, ErrProvider)
}

func TestLocalSessionSerializesProviderExecution(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	executor := &fakeExecutor{generateFn: func(context.Context, Request) (Response, error) {
		entered <- struct{}{}
		<-release

		return Response{Text: "ok"}, nil
	}}
	session, err := NewLocalSession(context.Background(), testModel(executor), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	errs := make(chan error, 2)
	for _, input := range []string{"first", "second"} {
		input := input
		go func() {
			_, err := Execute(context.Background(), session, OperationRequest{Mode: ModeGenerate, Input: input})
			errs <- err
		}()
	}

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first provider call did not start")
	}
	select {
	case <-entered:
		t.Fatal("second provider call overlapped the first")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)

	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	if len(session.History()) != 4 {
		t.Fatalf("expected two committed turns, got %#v", session.History())
	}
}

func TestLocalSessionTimeoutDoesNotCommit(t *testing.T) {
	executor := &fakeExecutor{generateFn: func(ctx context.Context, _ Request) (Response, error) {
		<-ctx.Done()

		return Response{}, ctx.Err()
	}}
	session, err := NewLocalSession(context.Background(), testModel(executor), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = Execute(context.Background(), session, OperationRequest{
		Mode:      ModeGenerate,
		Input:     "prompt",
		Execution: ExecutionOptions{Timeout: time.Millisecond},
	})
	requireCode(t, err, ErrTimeout)
	if len(session.History()) != 0 {
		t.Fatal("timed-out request mutated history")
	}
}

func TestLocalSessionCloseIsIdempotent(t *testing.T) {
	session, err := NewLocalSession(context.Background(), testModel(&fakeExecutor{}), SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = session.Generate(context.Background(), Request{})
	if err == nil || errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected closed-session error: %v", err)
	}
}
