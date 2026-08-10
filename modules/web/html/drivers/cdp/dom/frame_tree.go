package dom

import "github.com/mafredri/cdp/protocol/page"

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
