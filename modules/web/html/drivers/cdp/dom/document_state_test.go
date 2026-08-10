package dom

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestHTMLDocumentStateReplacement(t *testing.T) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("frame", "https://old.example"))
	oldState := &documentState{
		frameTree:  testFrameTree("frame", "https://old.example"),
		generation: 1,
	}
	oldElement := testStateElement(doc, oldState, "old-root")
	oldState.element = oldElement
	doc.replaceState(oldState)
	originalHash := doc.Hash()

	newState := &documentState{
		frameTree:  testFrameTree("frame", "https://new.example"),
		generation: 2,
	}
	newElement := testStateElement(doc, newState, "new-root")
	newState.element = newElement
	doc.replaceState(newState)

	if doc.Hash() != originalHash {
		t.Fatal("document identity changed with live state")
	}
	if got := doc.GetURL().String(); got != "https://new.example" {
		t.Fatalf("URL = %q, want replacement metadata", got)
	}
	if doc.GetElement() != newElement {
		t.Fatal("document did not expose the replacement root")
	}
	if _, err := oldElement.GetValue(context.Background()); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("old element error = %v, want detached", err)
	}
	if _, err := oldElement.AsAttributeTarget().GetAttribute(context.Background(), "id"); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("old capability error = %v, want detached", err)
	}
	if err := oldElement.Click(context.Background(), 1); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("old direct operation error = %v, want detached", err)
	}

	if _, err := oldElement.Get(context.Background(), runtime.None); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("old local operation error = %v, want detached", err)
	}
}

func TestDetachedDocumentRetainsMetadata(t *testing.T) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("child", "https://frame.example"))
	state := &documentState{frameTree: testFrameTree("child", "https://frame.example"), generation: 1}
	element := testStateElement(doc, state, "root")
	state.element = element
	doc.replaceState(state)
	doc.detach()

	if doc.currentState() != nil {
		t.Fatal("detached document retained active state")
	}

	if got := doc.GetURL().String(); got != "https://frame.example" {
		t.Fatalf("URL = %q", got)
	}

	if doc.GetElement() != element {
		t.Fatal("detached document did not retain its last root metadata")
	}

	if err := element.executor.run(context.Background(), func() error { return nil }); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("detached root error = %v, want detached", err)
	}

	if _, err := doc.GetCurrentURL(context.Background()); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("operation error = %v, want detached", err)
	}
}

func TestDocumentOperationContinuesAfterReplacement(t *testing.T) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("frame", "old"))
	doc.replaceState(&documentState{frameTree: testFrameTree("frame", "old"), generation: 1})

	calls := 0
	value, err := withDocumentResult(context.Background(), doc, func(state *documentState) (runtime.String, error) {
		calls++
		if calls == 1 {
			doc.replaceState(&documentState{frameTree: testFrameTree("frame", "new"), generation: 2})
			return runtime.EmptyString, drivers.ErrDetached
		}

		frame, err := doc.snapshotFrame()
		if err != nil {
			return runtime.EmptyString, err
		}

		return runtime.NewString(frame.Frame.URL), nil
	})
	if err != nil {
		t.Fatalf("operation: %v", err)
	}
	if value != "new" || calls != 2 {
		t.Fatalf("value = %q, calls = %d", value, calls)
	}
}

type refreshPageAPI struct {
	cdp.Page
	frameStarted chan struct{}
	frameRelease chan struct{}
	frameTree    page.FrameTree
	frameCalls   int
	mu           sync.Mutex
	blockOnFrame bool
}

func (api *refreshPageAPI) GetFrameTree(ctx context.Context) (*page.GetFrameTreeReply, error) {
	api.mu.Lock()
	api.frameCalls++
	api.mu.Unlock()

	if api.frameStarted != nil {
		close(api.frameStarted)
	}

	if api.frameRelease != nil {
		select {
		case <-api.frameRelease:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if api.blockOnFrame {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	frameTree := api.frameTree
	if frameTree.Frame.ID == "" {
		frameTree = testFrameTree("frame", "https://new.example")
	}

	return &page.GetFrameTreeReply{FrameTree: frameTree}, nil
}

func (api *refreshPageAPI) CreateIsolatedWorld(
	context.Context,
	*page.CreateIsolatedWorldArgs,
) (*page.CreateIsolatedWorldReply, error) {
	return &page.CreateIsolatedWorldReply{ExecutionContextID: 2}, nil
}

type refreshRuntimeAPI struct {
	cdp.Runtime
}

func (api *refreshRuntimeAPI) CallFunctionOn(
	context.Context,
	*cdpruntime.CallFunctionOnArgs,
) (*cdpruntime.CallFunctionOnReply, error) {
	id := cdpruntime.RemoteObjectID("new-root")

	return &cdpruntime.CallFunctionOnReply{Result: cdpruntime.RemoteObject{Type: "object", ObjectID: &id}}, nil
}

func TestDocumentRefreshIsSerializedByGeneration(t *testing.T) {
	doc, state, pageAPI := testRefreshDocument(t, false)

	const callers = 12
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- doc.refresh(context.Background(), state)
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("refresh: %v", err)
		}
	}
	if pageAPI.frameCalls != 1 {
		t.Fatalf("frame reloads = %d, want 1", pageAPI.frameCalls)
	}
	current := doc.currentState()
	if current == nil || current.generation != 2 || !doc.isCurrentState(current) {
		t.Fatal("replacement generation was not installed")
	}
}

func TestDocumentRefreshHonorsCancellation(t *testing.T) {
	doc, state, _ := testRefreshDocument(t, true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := doc.refresh(ctx, state); !errors.Is(err, context.Canceled) {
		t.Fatalf("refresh error = %v, want cancellation", err)
	}
}

func TestDocumentRefreshDoesNotOverwriteConcurrentReplacement(t *testing.T) {
	doc, observedState, pageAPI := testRefreshDocument(t, false)
	pageAPI.frameStarted = make(chan struct{})
	pageAPI.frameRelease = make(chan struct{})
	errCh := make(chan error, 1)

	go func() {
		errCh <- doc.refresh(context.Background(), observedState)
	}()

	<-pageAPI.frameStarted
	replacement := &documentState{
		frameTree:  testFrameTree("frame", "https://replacement.example"),
		generation: 2,
	}
	doc.replaceState(replacement)
	close(pageAPI.frameRelease)

	if err := <-errCh; err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if doc.currentState() != replacement {
		t.Fatal("refresh overwrote a concurrent replacement")
	}

	if got := doc.GetURL().String(); got != "https://replacement.example" {
		t.Fatalf("URL = %q, want replacement URL", got)
	}
}

func TestRootDocumentRefreshReplacesRetainedFrameStates(t *testing.T) {
	oldTree := testFrameTree("root", "https://old.example")
	oldChildTree := testFrameTree("child", "https://old-frame.example")
	oldTree.ChildFrames = []page.FrameTree{oldChildTree}

	newTree := testFrameTree("root", "https://new.example")
	newChildTree := testFrameTree("child", "https://new-frame.example")
	newTree.ChildFrames = []page.FrameTree{newChildTree}

	pageAPI := &refreshPageAPI{frameTree: newTree}
	client := &cdp.Client{Page: pageAPI, Runtime: new(refreshRuntimeAPI)}
	manager, err := New(zerolog.Nop(), client, nil, nil)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}

	root := newHTMLDocument(zerolog.Nop(), manager, oldTree)
	rootState := &documentState{frameTree: oldTree, generation: 1}
	root.replaceState(rootState)
	child := newHTMLDocument(zerolog.Nop(), manager, oldChildTree)
	childState := &documentState{
		frameTree:  oldChildTree,
		generation: 1,
	}
	oldChildElement := testStateElement(child, childState, "old-child-root")
	childState.element = oldChildElement
	child.replaceState(childState)
	manager.SetMainFrame(root)
	manager.frames.Set("child", Frame{tree: oldChildTree, node: child})

	if err := root.refresh(context.Background(), rootState); err != nil {
		t.Fatalf("refresh root: %v", err)
	}
	rootCurrent := root.currentState()
	childCurrent := child.currentState()
	if rootCurrent == nil || rootCurrent.generation != 2 || childCurrent == nil || childCurrent.generation != 2 {
		t.Fatal("root refresh did not replace every retained document generation")
	}
	if got := child.GetURL().String(); got != "https://new-frame.example" {
		t.Fatalf("child URL = %q, want replacement metadata", got)
	}
	if _, err := oldChildElement.GetValue(context.Background()); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("old child element error = %v, want detached", err)
	}
}

func TestDocumentStaleOperationRetriesOnce(t *testing.T) {
	doc, _, pageAPI := testRefreshDocument(t, false)
	calls := 0
	stale := &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed."}

	_, err := withDocumentResult(context.Background(), doc, func(*documentState) (runtime.Value, error) {
		calls++
		return runtime.None, stale
	})
	if !errors.Is(err, stale) {
		t.Fatalf("operation error = %v, want stale error", err)
	}
	if calls != 2 {
		t.Fatalf("operation calls = %d, want 2", calls)
	}
	if pageAPI.frameCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", pageAPI.frameCalls)
	}
}

func TestDocumentOperationDoesNotRefreshUnrelatedError(t *testing.T) {
	doc, _, pageAPI := testRefreshDocument(t, false)
	want := errors.New("unrelated protocol failure")
	calls := 0

	_, err := withDocumentResult(context.Background(), doc, func(*documentState) (runtime.Value, error) {
		calls++

		return runtime.None, want
	})
	if !errors.Is(err, want) {
		t.Fatalf("operation error = %v, want %v", err, want)
	}

	if calls != 1 {
		t.Fatalf("operation calls = %d, want 1", calls)
	}

	if pageAPI.frameCalls != 0 {
		t.Fatalf("refresh calls = %d, want 0", pageAPI.frameCalls)
	}
}

func TestElementExecutorRunsWhileStateCurrent(t *testing.T) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("frame", "https://example.com"))
	state := &documentState{frameTree: testFrameTree("frame", "https://example.com"), generation: 1}
	element := testStateElement(doc, state, "root")
	state.element = element
	doc.replaceState(state)

	called := false
	err := element.executor.run(context.Background(), func() error {
		called = true

		return nil
	})
	if err != nil {
		t.Fatalf("run current element operation: %v", err)
	}

	if !called {
		t.Fatal("current element operation was not executed")
	}
}

func TestElementExecutorNormalizesStaleErrors(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{name: "context", message: "Execution context was destroyed."},
		{name: "object", message: "Could not find object with given id"},
		{name: "node", message: "Node with given id does not belong to the document"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, state, pageAPI := testRefreshDocument(t, false)
			element := testStateElement(doc, state, "old-root")
			state.element = element
			doc.replaceState(state)
			stale := &rpcc.ResponseError{Code: -32000, Message: tt.message}
			calls := 0

			err := element.executor.run(context.Background(), func() error {
				calls++

				return stale
			})
			if !errors.Is(err, drivers.ErrDetached) {
				t.Fatalf("stale operation error = %v, want detached", err)
			}

			if calls != 1 || pageAPI.frameCalls != 1 {
				t.Fatalf("operation calls = %d, refresh calls = %d; want 1 each", calls, pageAPI.frameCalls)
			}

			if doc.isCurrentState(state) {
				t.Fatal("stale element state remained current after refresh")
			}

			err = element.executor.run(context.Background(), func() error {
				calls++

				return nil
			})
			if !errors.Is(err, drivers.ErrDetached) {
				t.Fatalf("old element retry error = %v, want detached", err)
			}

			if calls != 1 {
				t.Fatalf("old element was rebound and executed %d operations", calls)
			}
		})
	}
}

func TestElementExecutorPropagatesRefreshFailure(t *testing.T) {
	doc, state, _ := testRefreshDocument(t, true)
	element := testStateElement(doc, state, "root")
	state.element = element
	doc.replaceState(state)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stale := &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed."}

	err := element.executor.run(ctx, func() error { return stale })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("refresh failure = %v, want cancellation", err)
	}
}

func TestManagedElementExecutorPreservesUnrelatedError(t *testing.T) {
	doc, state, pageAPI := testRefreshDocument(t, false)
	element := testStateElement(doc, state, "root")
	state.element = element
	doc.replaceState(state)
	want := errors.New("unrelated protocol failure")

	err := element.executor.run(context.Background(), func() error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("operation error = %v, want %v", err, want)
	}

	if pageAPI.frameCalls != 0 {
		t.Fatalf("refresh calls = %d, want 0", pageAPI.frameCalls)
	}
}

func TestUnmanagedElementExecutorPreservesStaleError(t *testing.T) {
	executor := newElementExecutor(nil)
	want := &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed."}
	calls := 0

	err := executor.run(context.Background(), func() error {
		calls++

		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("operation error = %v, want %v", err, want)
	}

	if calls != 1 {
		t.Fatalf("operation calls = %d, want 1", calls)
	}
}

func TestFrameTreeUpdatePreservesStateDuringReplacement(t *testing.T) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("frame", "https://old.example"))
	oldState := &documentState{frameTree: testFrameTree("frame", "https://old.example"), generation: 1}
	oldElement := testStateElement(doc, oldState, "old-root")
	oldState.element = oldElement
	doc.replaceState(oldState)

	doc.updateFrameTree(testFrameTree("frame", "https://metadata.example"))
	if doc.currentState() != oldState || oldState.generation != 1 {
		t.Fatal("metadata update replaced the document state or changed its generation")
	}

	newState := &documentState{frameTree: testFrameTree("frame", "https://new.example"), generation: 2}
	newElement := testStateElement(doc, newState, "new-root")
	newState.element = newElement
	started := make(chan struct{})
	errs := make(chan error, 2000)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		close(started)

		for i := range 1000 {
			doc.updateFrameTree(testFrameTree("frame", runtime.NewInt(i).String()))
		}
	}()

	<-started
	wg.Add(1)
	go func() {
		defer wg.Done()

		for range 1000 {
			if _, err := doc.snapshotFrame(); err != nil {
				errs <- err
			}

			if _, err := withDocumentResult(context.Background(), doc, func(state *documentState) (uint64, error) {
				return state.generation, nil
			}); err != nil {
				errs <- err
			}
		}
	}()

	doc.replaceState(newState)
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent document operation: %v", err)
	}

	if doc.currentState() != newState || newState.generation != 2 {
		t.Fatal("frame metadata update produced a stale duplicate state")
	}

	if err := oldElement.executor.run(context.Background(), func() error { return nil }); !errors.Is(err, drivers.ErrDetached) {
		t.Fatalf("old element error = %v, want detached", err)
	}

	if err := newElement.executor.run(context.Background(), func() error { return nil }); err != nil {
		t.Fatalf("new element operation: %v", err)
	}
}

func testRefreshDocument(t *testing.T, block bool) (*HTMLDocument, *documentState, *refreshPageAPI) {
	t.Helper()

	pageAPI := &refreshPageAPI{blockOnFrame: block}
	client := &cdp.Client{Page: pageAPI, Runtime: new(refreshRuntimeAPI)}
	manager, err := New(zerolog.Nop(), client, nil, nil)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	doc := newHTMLDocument(zerolog.Nop(), manager, testFrameTree("frame", "https://old.example"))
	state := &documentState{frameTree: testFrameTree("frame", "https://old.example"), generation: 1}
	doc.replaceState(state)
	manager.SetMainFrame(doc)

	return doc, state, pageAPI
}

func testStateElement(doc *HTMLDocument, state *documentState, id cdpruntime.RemoteObjectID) *HTMLElement {
	element := &HTMLElement{id: id}
	element.installExecutor(newStateElementExecutor(doc, state))

	return element
}

func testFrameTree(id page.FrameID, url string) page.FrameTree {
	return page.FrameTree{Frame: page.Frame{ID: id, URL: url}}
}

func BenchmarkElementExecutorLifecycle(b *testing.B) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("frame", "https://example.com"))
	state := &documentState{
		frameTree:  testFrameTree("frame", "https://example.com"),
		generation: 1,
	}
	element := testStateElement(doc, state, "root")
	state.element = element
	doc.replaceState(state)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := element.executor.run(ctx, func() error { return nil }); err != nil {
			b.Fatal(err)
		}
	}
}
