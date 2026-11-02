package dom

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/synctest"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func TestRecoveredDocumentMetadata(t *testing.T) {
	for _, operation := range []string{"load root", "reload new root", "reload root", "refresh root", "refresh child", "lazy child", "frame list"} {
		t.Run(operation, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				failed := false
				manager, pageAPI, _, trace := newLoadingTestManager(t, func(_ context.Context, stage string) error {
					if stage == "eval" && !failed {
						failed = true

						return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
					}

					return nil
				})
				old := testFrameTree("frame", "https://old.example")
				old.Frame.LoaderID = "old"
				fresh := testFrameTree("frame", "https://new.example")
				fresh.Frame.LoaderID = "new"
				isChild := operation == "refresh child" || operation == "lazy child" || operation == "frame list"
				if isChild {
					old.Frame.ID, fresh.Frame.ID = "child", "child"
				}

				pageAPI.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
					tree := old
					if failed {
						tree = fresh
					}

					if isChild {
						root := testFrameTree("frame", "https://root.example")
						root.ChildFrames = []page.FrameTree{tree}
						tree = root
					}

					return &page.GetFrameTreeReply{FrameTree: tree}, nil
				}
				var doc *HTMLDocument
				var previous *documentState
				if operation == "reload root" || operation == "refresh root" || operation == "refresh child" {
					seed := old
					seed.Frame.LoaderID = "seed"
					doc = seedLoadingDocument(manager, seed)
					previous = doc.currentState()
					if !isChild {
						manager.SetMainFrame(doc)
					} else {
						manager.frames.Set("child", Frame{tree: seed, node: doc})
					}
				}

				var err error
				switch operation {
				case "load root":
					doc, err = manager.LoadRootDocument(t.Context())
				case "reload new root", "reload root":
					err = manager.ReloadRoot(t.Context())
					doc = manager.GetMainFrame()
				case "refresh root", "refresh child":
					err = doc.refresh(t.Context(), previous)
				case "lazy child", "frame list":
					manager.AddFrame(old)
					if operation == "frame list" {
						_, err = manager.GetFrameNodes(t.Context())
					}

					if err == nil {
						doc, err = manager.GetFrameNode(t.Context(), "child")
					}
				}

				if err != nil {
					t.Fatal(err)
				}

				if got := doc.Frame().Frame; got.LoaderID != "new" || got.URL != fresh.Frame.URL {
					t.Fatalf("recovered document metadata: %+v", got)
				}

				if operation != "load root" {
					cached, err := manager.GetFrameTree(t.Context(), fresh.Frame.ID)
					if err != nil || cached.Frame.LoaderID != "new" || cached.Frame.URL != fresh.Frame.URL {
						t.Fatalf("cached frame=%+v, error=%v", cached.Frame, err)
					}
				}

				if previous != nil {
					if doc.currentState().generation != previous.generation+1 {
						t.Fatal("failed attempt advanced the generation")
					}

					if _, err := previous.element.GetValue(t.Context()); !errors.Is(err, drivers.ErrDetached) {
						t.Fatalf("old element remained attached: %v", err)
					}
				}

				if len(trace.contexts) != 2 || trace.contexts[0] == trace.contexts[1] {
					t.Fatalf("execution contexts: %v", trace.contexts)
				}
			})
		})
	}
}

func TestDocumentRecoveryReselectsFrameClient(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager, rootAPI, _, _ := newLoadingTestManager(t, nil)
		oldFrame := testFrameTree("child", "https://old.example")
		fresh := testFrameTree("child", "https://new.example")
		fresh.Frame.LoaderID = "new"
		rootCalls, replacementCalls := 0, 0
		replacement := &cdp.Client{
			Page: &loadingPageAPI{createWorld: func(_ context.Context, args *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
				if args.FrameID != "child" {
					t.Fatal("frame identity changed")
				}

				replacementCalls++

				return &page.CreateIsolatedWorldReply{ExecutionContextID: 99}, nil
			}},
			Runtime: &loadingRuntimeAPI{call: func(_ context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
				if *args.ExecutionContextID != 99 {
					t.Fatal("used the old client's context")
				}

				replacementCalls++

				return loadingRootReply(), nil
			}},
		}
		oldClient := &cdp.Client{Page: &loadingPageAPI{createWorld: func(context.Context, *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
			rootCalls++
			manager.RecordFrameClient("child", replacement)

			return nil, &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
		}}}
		manager.RecordFrameClient("child", oldClient)
		rootAPI.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
			root := testFrameTree("root", "https://root.example")
			root.ChildFrames = []page.FrameTree{fresh}

			return &page.GetFrameTreeReply{FrameTree: root}, nil
		}

		doc, err := manager.LoadDocument(t.Context(), oldFrame)
		if err != nil {
			t.Fatal(err)
		}

		if doc.currentState().client != replacement || rootCalls != 1 || replacementCalls != 2 || doc.Frame().Frame.LoaderID != "new" {
			t.Fatal("recovery did not keep client, execution context, and frame metadata together")
		}
	})
}

func TestDocumentRecoveryFrameRemoval(t *testing.T) {
	for _, operation := range []string{"load", "refresh", "reload"} {
		for _, replace := range []bool{false, true} {
			if operation == "load" && replace {
				continue
			}

			t.Run(fmt.Sprintf("%s/replaced=%t", operation, replace), func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					failed := false
					manager, pageAPI, _, trace := newLoadingTestManager(t, func(_ context.Context, stage string) error {
						if stage == "eval" {
							failed = true

							return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
						}

						return nil
					})
					frame := testFrameTree("frame", "https://old.example")
					doc := seedLoadingDocument(manager, frame)
					observed := doc.currentState()
					replacement := &documentState{frameTree: testFrameTree("frame", "https://replacement.example"), generation: 2}
					pageAPI.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
						tree := frame
						if failed {
							tree = testFrameTree("another-frame", "https://elsewhere.example")
							if replace {
								doc.replaceState(replacement)
							}
						}

						return &page.GetFrameTreeReply{FrameTree: tree}, nil
					}
					var err error
					switch operation {
					case "load":
						_, err = manager.LoadDocument(t.Context(), frame)
					case "refresh":
						err = doc.refresh(t.Context(), observed)
					case "reload":
						_, err = doc.reload(t.Context(), frame)
					}

					if replace {
						if err != nil || doc.currentState() != replacement {
							t.Fatalf("concurrent replacement lost: %v", err)
						}
					} else if !errors.Is(err, drivers.ErrDetached) {
						t.Fatalf("error=%v, want detached", err)
					} else if operation != "load" && doc.currentState() != nil {
						t.Fatal("removed document remained attached")
					}

					if len(trace.contexts) != 1 {
						t.Fatal("removed frame was evaluated again")
					}
				})
			})
		}
	}
}

func seedLoadingDocument(manager *Manager, frame page.FrameTree) *HTMLDocument {
	doc := newHTMLDocument(zerolog.Nop(), manager, frame)
	state := &documentState{frameTree: frame, generation: 1}
	state.element = testStateElement(doc, state, "old-root")
	doc.replaceState(state)

	return doc
}

func TestRecoveredChildSubtreeSurvivesRootInstallation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager, pageAPI, runtimeAPI, _ := newLoadingTestManager(t, nil)
		rootTree := testFrameTree("frame", "https://root.example")
		rootTree.Frame.LoaderID = "next-root"
		childTree := testFrameTree("child", "https://old-child.example")
		childTree.Frame.LoaderID = "old-child"
		leafTree := testFrameTree("leaf", "https://old-leaf.example")
		removedTree := testFrameTree("removed", "https://removed.example")
		childTree.ChildFrames = []page.FrameTree{leafTree, removedTree}
		rootTree.ChildFrames = []page.FrameTree{childTree}
		freshChild := testFrameTree("child", "https://new-child.example")
		freshChild.Frame.LoaderID = "new-child"
		freshLeaf := testFrameTree("leaf", "https://new-leaf.example")
		freshLeaf.Frame.LoaderID = "new-leaf"
		freshChild.ChildFrames = []page.FrameTree{freshLeaf, testFrameTree("added", "https://added.example")}
		freshRoot := rootTree
		freshRoot.ChildFrames = []page.FrameTree{freshChild}

		seed := rootTree
		seed.Frame.LoaderID = "seed-root"
		root := seedLoadingDocument(manager, seed)
		child := seedLoadingDocument(manager, childTree)
		leaf := seedLoadingDocument(manager, leafTree)
		removed := seedLoadingDocument(manager, removedTree)
		oldLeaf := leaf.currentState().element
		manager.SetMainFrame(root)
		for _, doc := range []*HTMLDocument{child, leaf, removed} {
			manager.frames.Set(doc.identity, Frame{tree: doc.Frame(), node: doc})
		}

		worlds := map[cdpruntime.ExecutionContextID]page.FrameID{}
		creates := map[page.FrameID]int{}
		failed := false
		pageAPI.createWorld = func(_ context.Context, args *page.CreateIsolatedWorldArgs) (*page.CreateIsolatedWorldReply, error) {
			id := cdpruntime.ExecutionContextID(len(worlds) + 1)
			worlds[id] = args.FrameID
			creates[args.FrameID]++

			return &page.CreateIsolatedWorldReply{ExecutionContextID: id}, nil
		}
		pageAPI.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
			tree := rootTree
			if failed {
				tree = freshRoot
			}

			return &page.GetFrameTreeReply{FrameTree: tree}, nil
		}
		runtimeAPI.call = func(_ context.Context, args *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
			if worlds[*args.ExecutionContextID] == "child" && !failed {
				failed = true

				return nil, &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
			}

			return loadingRootReply(), nil
		}

		if err := manager.ReloadRoot(t.Context()); err != nil {
			t.Fatal(err)
		}

		if creates["frame"] != 1 || creates["child"] != 2 || creates["leaf"] != 1 || creates["removed"] != 0 {
			t.Fatalf("unexpected document acquisitions: %v", creates)
		}

		for _, want := range []page.FrameTree{freshChild, freshLeaf} {
			cached, err := manager.GetFrameTree(t.Context(), want.Frame.ID)
			if err != nil || cached.Frame.LoaderID != want.Frame.LoaderID || cached.Frame.URL != want.Frame.URL {
				t.Fatalf("cached frame=%+v, error=%v", cached.Frame, err)
			}

			fromRoot, found := findFrame(root.Frame(), want.Frame.ID)
			if !found || fromRoot.Frame.LoaderID != want.Frame.LoaderID {
				t.Fatal("root installation overwrote recovered descendant metadata")
			}
		}

		if manager.frames.Has("removed") || removed.currentState() != nil || !manager.frames.Has("added") {
			t.Fatal("recovered subtree membership was lost")
		}

		if child.currentState().generation != 2 || leaf.currentState().generation != 2 {
			t.Fatal("recovery published an extra generation")
		}

		if _, err := oldLeaf.GetValue(t.Context()); !errors.Is(err, drivers.ErrDetached) {
			t.Fatalf("old descendant element remained attached: %v", err)
		}

		if rootTree.ChildFrames[0].Frame.LoaderID != "old-child" || childTree.ChildFrames[0].Frame.URL != leafTree.Frame.URL {
			t.Fatal("recovery mutated an earlier frame snapshot")
		}
	})
}

func TestDocumentRecoveryPreservesConcurrentReplacement(t *testing.T) {
	for _, operation := range []string{"refresh", "reload"} {
		t.Run(operation, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				var doc *HTMLDocument
				replacement := &documentState{frameTree: testFrameTree("frame", "https://replacement.example"), generation: 2}
				manager, _, _, _ := newLoadingTestManager(t, func(_ context.Context, stage string) error {
					if stage != "eval" {
						return nil
					}

					calls++
					if calls == 1 {
						return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
					}

					doc.replaceState(replacement)

					return nil
				})
				doc = seedLoadingDocument(manager, testFrameTree("frame", "old"))
				manager.SetMainFrame(doc)
				identity := doc.Hash()

				var err error
				if operation == "refresh" {
					err = doc.refresh(t.Context(), doc.currentState())
				} else {
					_, err = doc.reload(t.Context(), doc.Frame())
				}

				if err != nil || doc.currentState() != replacement || doc.Hash() != identity || calls != 2 {
					t.Fatalf("replacement lost or identity changed: %v", err)
				}
			})
		})
	}
}

func TestDocumentRecoverySerializesRefresh(t *testing.T) {
	failed := false
	manager, _, _, trace := newLoadingTestManager(t, func(_ context.Context, stage string) error {
		if stage == "eval" && !failed {
			failed = true

			return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
		}

		return nil
	})
	doc := seedLoadingDocument(manager, testFrameTree("frame", "old"))
	manager.SetMainFrame(doc)
	observed := doc.currentState()
	const callers = 12
	finished := make(chan error, callers)

	for range callers {
		go func() { finished <- doc.refresh(t.Context(), observed) }()
	}

	for range callers {
		if err := <-finished; err != nil {
			t.Fatal(err)
		}
	}

	if doc.currentState().generation != 2 || !slices.Equal(trace.stages, []string{"tree", "create", "eval", "tree", "create", "eval"}) {
		t.Fatalf("refresh was repeated across callers: %v", trace.stages)
	}
}

func TestDocumentRecoveryFailurePreservesState(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		stale := &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
		var doc *HTMLDocument
		var observed *documentState
		manager, _, _, _ := newLoadingTestManager(t, func(_ context.Context, stage string) error {
			if doc.currentState() != observed {
				t.Fatal("published state before acquisition succeeded")
			}

			if stage == "eval" {
				return stale
			}

			return nil
		})
		doc = seedLoadingDocument(manager, testFrameTree("frame", "old"))
		observed = doc.currentState()

		if err := doc.refresh(t.Context(), observed); !errors.Is(err, stale) {
			t.Fatalf("error=%v, want stale error", err)
		}

		if doc.currentState() != observed || observed.generation != 1 || doc.Frame().Frame.URL != "old" {
			t.Fatal("failed acquisition changed the document")
		}
	})
}

func TestRetainedFrameReloadPreservesNewerRoot(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager, pageAPI, runtimeAPI, _ := newLoadingTestManager(t, nil)
		tree := testFrameTree("frame", "https://loading.example")
		childTree := testFrameTree("child", "https://child.example")
		tree.ChildFrames = []page.FrameTree{childTree}
		root := seedLoadingDocument(manager, tree)
		child := seedLoadingDocument(manager, childTree)
		manager.SetMainFrame(root)
		manager.frames.Set("child", Frame{tree: childTree, node: child})
		replacement := &documentState{frameTree: testFrameTree("frame", "https://replacement.example"), generation: 3}
		pageAPI.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
			return &page.GetFrameTreeReply{FrameTree: tree}, nil
		}
		calls := 0
		runtimeAPI.call = func(context.Context, *cdpruntime.CallFunctionOnArgs) (*cdpruntime.CallFunctionOnReply, error) {
			calls++
			if calls == 2 {
				return nil, &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
			}

			if calls == 3 {
				root.replaceState(replacement)
			}

			return loadingRootReply(), nil
		}

		if err := manager.ReloadRoot(t.Context()); err != nil {
			t.Fatal(err)
		}

		if calls != 3 || root.currentState() != replacement || root.Frame().Frame.URL != "https://replacement.example" {
			t.Fatal("installing the recovered subtree overwrote a newer root generation")
		}
	})
}
