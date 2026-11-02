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
	"github.com/mafredri/cdp/rpcc"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func TestLazyFrameReconcilesDescendants(t *testing.T) {
	for _, entry := range []string{"child documents", "all frames"} {
		t.Run(entry, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				stale := true
				manager, api, _, _ := newLoadingTestManager(t, func(_ context.Context, stage string) error {
					if stage == "eval" && stale {
						stale = false

						return &rpcc.ResponseError{Code: -32000, Message: "Cannot find context with specified id"}
					}

					return nil
				})
				retained := testFrameTree("retained", "retained")
				removed := testFrameTree("removed", "removed")
				old := testFrameTree("parent", "old")
				old.ChildFrames = []page.FrameTree{removed, retained}
				ancestor := testFrameTree("ancestor", "ancestor")
				ancestor.ChildFrames = []page.FrameTree{old}
				outside := testFrameTree("outside", "outside")
				rootTree := testFrameTree("root", "root")
				rootTree.ChildFrames = []page.FrameTree{ancestor, outside}
				root := seedLoadingDocument(manager, rootTree)
				manager.SetMainFrame(root)
				docs := map[page.FrameID]*HTMLDocument{}
				for _, tree := range []page.FrameTree{ancestor, retained, removed, outside} {
					doc := seedLoadingDocument(manager, tree)
					docs[tree.Frame.ID] = doc
					manager.frames.Set(doc.identity, Frame{tree: tree, node: doc})
					manager.RecordFrameClient(doc.identity, manager.rootClient)
				}

				added := testFrameTree("added", "added")
				added.ChildFrames = []page.FrameTree{testFrameTree("leaf", "leaf")}
				fresh := testFrameTree("parent", "fresh")
				fresh.ChildFrames = []page.FrameTree{retained, added}
				freshAncestor := ancestor
				freshAncestor.ChildFrames = []page.FrameTree{fresh}
				freshRoot := rootTree
				freshRoot.ChildFrames = []page.FrameTree{freshAncestor, outside}
				api.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
					return &page.GetFrameTreeReply{FrameTree: freshRoot}, nil
				}

				var list runtime.List
				var err error
				var want []page.FrameID
				if entry == "all frames" {
					list, err = manager.GetFrameNodes(t.Context())
					want = []page.FrameID{"added", "ancestor", "leaf", "outside", "parent", "retained", "root"}
				} else {
					var parent *HTMLDocument
					parent, err = manager.GetFrameNode(t.Context(), "parent")
					if err != nil {
						t.Fatal(err)
					}

					list, err = parent.GetChildDocuments(t.Context())
					want = []page.FrameID{"added", "retained"}
				}

				if err != nil {
					t.Fatal(err)
				}

				n, err := list.Length(t.Context())
				if err != nil {
					t.Fatal(err)
				}

				var ids []page.FrameID
				for i := range n {
					v, err := list.At(t.Context(), i)
					if err != nil {
						t.Fatal(err)
					}

					ids = append(ids, v.(*HTMLDocument).identity)
				}

				slices.Sort(ids)
				if !slices.Equal(ids, want) {
					t.Fatalf("frame IDs %v, want %v", ids, want)
				}

				for _, doc := range []*HTMLDocument{root, docs["ancestor"]} {
					cached, err := manager.GetFrameTree(t.Context(), doc.identity)
					if err != nil {
						t.Fatal(err)
					}

					for _, tree := range []page.FrameTree{doc.Frame(), cached} {
						if _, ok := findFrame(tree, "leaf"); !ok {
							t.Fatalf("ancestor %s is missing the recovered leaf", doc.identity)
						}
					}
				}

				for _, id := range []page.FrameID{"retained", "outside"} {
					doc, err := manager.GetFrameNode(t.Context(), id)
					if err != nil || doc != docs[id] || doc.currentState() == nil {
						t.Fatalf("lost retained document %s", id)
					}
				}

				if _, err := manager.GetFrameNode(t.Context(), "removed"); !errors.Is(err, runtime.ErrNotFound) {
					t.Fatalf("removed frame lookup: %v", err)
				}

				if _, err := docs["removed"].snapshot(); !errors.Is(err, drivers.ErrDetached) {
					t.Fatalf("removed document: %v", err)
				}

				if _, found := manager.owners.Get("removed"); found {
					t.Fatal("removed frame retained its owner")
				}

				if _, found := findFrame(rootTree, "removed"); !found {
					t.Fatal("mutated the old root snapshot")
				}
			})
		})
	}
}

func BenchmarkGetFrameNodes(b *testing.B) {
	manager, err := New(zerolog.Nop(), new(cdp.Client), nil, nil)
	if err != nil {
		b.Fatal(err)
	}

	tree := testFrameTree("root", "root")
	for i := range 16 {
		id := page.FrameID(fmt.Sprintf("child-%d", i))
		tree.ChildFrames = append(tree.ChildFrames, testFrameTree(id, string(id)))
	}

	manager.SetMainFrame(seedLoadingDocument(manager, tree))
	for _, child := range tree.ChildFrames {
		manager.frames.Set(child.Frame.ID, Frame{tree: child, node: seedLoadingDocument(manager, child)})
	}

	ctx := b.Context()
	for b.Loop() {
		if _, err := manager.GetFrameNodes(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func TestGetFrameNodesIncludesIndependentTrees(t *testing.T) {
	manager, _, _, _ := newLoadingTestManager(t, nil)
	root := testFrameTree("root", "root")
	root.ChildFrames = []page.FrameTree{testFrameTree("removed", "removed")}
	manager.SetMainFrame(seedLoadingDocument(manager, root))

	if err := manager.RemoveFrame("removed"); err != nil {
		t.Fatal(err)
	}

	standalone := testFrameTree("independent", "independent")
	standalone.ChildFrames = []page.FrameTree{testFrameTree("child", "child")}
	manager.AddFrame(standalone)

	frames, err := manager.GetFrameNodes(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	var ids []page.FrameID
	err = frames.ForEach(t.Context(), func(_ context.Context, value runtime.Value, _ runtime.Int) (runtime.Boolean, error) {
		ids = append(ids, value.(*HTMLDocument).identity)

		return true, nil
	})
	if err != nil || !slices.Equal(ids, []page.FrameID{"root", "independent", "child"}) {
		t.Fatalf("frame IDs %v, error %v", ids, err)
	}
}

func TestLazyFrameRecoveryPreservesNewerAncestor(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		stale := true
		manager, api, _, _ := newLoadingTestManager(t, func(_ context.Context, stage string) error {
			if stage == "eval" && stale {
				stale = false

				return &rpcc.ResponseError{Code: -32000, Message: "Execution context was destroyed"}
			}

			return nil
		})
		tree := testFrameTree("root", "old-root")
		tree.ChildFrames = []page.FrameTree{testFrameTree("child", "old-child")}
		root := seedLoadingDocument(manager, tree)
		manager.SetMainFrame(root)
		api.frameTree = func(context.Context) (*page.GetFrameTreeReply, error) {
			root.replaceState(&documentState{frameTree: testFrameTree("root", "newer-root"), generation: 2})
			fresh := tree
			fresh.ChildFrames = []page.FrameTree{testFrameTree("child", "fresh-child")}

			return &page.GetFrameTreeReply{FrameTree: fresh}, nil
		}

		if _, err := manager.GetFrameNode(t.Context(), "child"); err != nil {
			t.Fatal(err)
		}

		if got := root.Frame().Frame.URL; got != "newer-root" {
			t.Fatalf("lazy recovery overwrote a newer root snapshot: %s", got)
		}
	})
}
