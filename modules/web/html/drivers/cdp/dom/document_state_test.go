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
	oldElement := testGenerationElement(doc, 1, "old-root")
	doc.replaceState(&documentState{
		element:    oldElement,
		frameTree:  testFrameTree("frame", "https://old.example"),
		generation: 1,
	})
	originalHash := doc.Hash()

	newElement := testGenerationElement(doc, 2, "new-root")
	doc.replaceState(&documentState{
		element:    newElement,
		frameTree:  testFrameTree("frame", "https://new.example"),
		generation: 2,
	})

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
}

func TestDetachedDocumentRetainsMetadata(t *testing.T) {
	doc := newHTMLDocument(zerolog.Nop(), nil, testFrameTree("child", "https://frame.example"))
	doc.replaceState(&documentState{frameTree: testFrameTree("child", "https://frame.example"), generation: 1})
	doc.detach()

	if got := doc.GetURL().String(); got != "https://frame.example" {
		t.Fatalf("URL = %q", got)
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

		return runtime.NewString(state.frameTree.Frame.URL), nil
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
	frameTree    page.FrameTree
	frameCalls   int
	mu           sync.Mutex
	blockOnFrame bool
}

func (api *refreshPageAPI) GetFrameTree(ctx context.Context) (*page.GetFrameTreeReply, error) {
	api.mu.Lock()
	api.frameCalls++
	api.mu.Unlock()

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
	doc, pageAPI := testRefreshDocument(t, false)

	const callers = 12
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- doc.refresh(context.Background(), 1)
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
	if !doc.isGenerationCurrent(2) {
		t.Fatal("replacement generation was not installed")
	}
}

func TestDocumentRefreshHonorsCancellation(t *testing.T) {
	doc, _ := testRefreshDocument(t, true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := doc.refresh(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("refresh error = %v, want cancellation", err)
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
	root.replaceState(&documentState{frameTree: oldTree, generation: 1})
	child := newHTMLDocument(zerolog.Nop(), manager, oldChildTree)
	oldChildElement := testGenerationElement(child, 1, "old-child-root")
	child.replaceState(&documentState{
		element:    oldChildElement,
		frameTree:  oldChildTree,
		generation: 1,
	})
	manager.SetMainFrame(root)
	manager.frames.Set("child", Frame{tree: oldChildTree, node: child})

	if err := root.refresh(context.Background(), 1); err != nil {
		t.Fatalf("refresh root: %v", err)
	}
	if !root.isGenerationCurrent(2) || !child.isGenerationCurrent(2) {
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
	doc, pageAPI := testRefreshDocument(t, false)
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

func testRefreshDocument(t *testing.T, block bool) (*HTMLDocument, *refreshPageAPI) {
	t.Helper()

	pageAPI := &refreshPageAPI{blockOnFrame: block}
	client := &cdp.Client{Page: pageAPI, Runtime: new(refreshRuntimeAPI)}
	manager, err := New(zerolog.Nop(), client, nil, nil)
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	doc := newHTMLDocument(zerolog.Nop(), manager, testFrameTree("frame", "https://old.example"))
	doc.replaceState(&documentState{frameTree: testFrameTree("frame", "https://old.example"), generation: 1})
	manager.SetMainFrame(doc)

	return doc, pageAPI
}

func testGenerationElement(doc *HTMLDocument, generation uint64, id cdpruntime.RemoteObjectID) *HTMLElement {
	executor := newElementExecutor(nil, doc, generation)

	return &HTMLElement{
		eval:       executor,
		id:         id,
		document:   doc,
		generation: generation,
		attributes: newElementAttributes(executor, id),
		wait:       newElementWait(executor, id),
	}
}

func testFrameTree(id page.FrameID, url string) page.FrameTree {
	return page.FrameTree{Frame: page.Frame{ID: id, URL: url}}
}
