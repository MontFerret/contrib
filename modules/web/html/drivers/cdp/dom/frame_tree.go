package dom

import (
	"slices"

	"github.com/mafredri/cdp/protocol/page"
)

func collectFrameIDs(root page.FrameTree) map[page.FrameID]struct{} {
	out := map[page.FrameID]struct{}{
		root.Frame.ID: {},
	}

	for _, child := range root.ChildFrames {
		for id := range collectFrameIDs(child) {
			out[id] = struct{}{}
		}
	}

	return out
}

func findFrame(root page.FrameTree, frameID page.FrameID) (page.FrameTree, bool) {
	if root.Frame.ID == frameID {
		return root, true
	}

	for _, child := range root.ChildFrames {
		if frame, found := findFrame(child, frameID); found {
			return frame, true
		}
	}

	return page.FrameTree{}, false
}

// Copy only the path to the replaced frame. Earlier snapshots may share child slices.
func replaceFrameSubtree(tree, replacement page.FrameTree) (page.FrameTree, bool) {
	if tree.Frame.ID == replacement.Frame.ID {
		return replacement, true
	}

	for i, child := range tree.ChildFrames {
		updated, found := replaceFrameSubtree(child, replacement)
		if found {
			tree.ChildFrames = slices.Clone(tree.ChildFrames)
			tree.ChildFrames[i] = updated

			return tree, true
		}
	}

	return tree, false
}
