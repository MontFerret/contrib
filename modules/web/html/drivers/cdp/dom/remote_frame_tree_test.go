package dom

import (
	"testing"

	"github.com/mafredri/cdp/protocol/page"
)

func TestMergeNestedRendererTrees(t *testing.T) {
	root := testFrameTree("root", "root")
	parent := testFrameTree("parent", "parent")
	child := testFrameTree("child", "child")
	child.Frame.ParentID = &parent.Frame.ID
	parent.Frame.ParentID = &root.Frame.ID
	root.ChildFrames = []page.FrameTree{testFrameTree("local", "local"), testFrameTree("parent", "old")}

	mergeRemoteFrames(&root, []page.FrameTree{child, parent})

	if len(root.ChildFrames) != 2 || root.ChildFrames[0].Frame.ID != "local" || len(root.ChildFrames[1].ChildFrames) != 1 || root.ChildFrames[1].ChildFrames[0].Frame.ID != "child" {
		t.Fatalf("merged frame hierarchy = %+v", root)
	}
}
