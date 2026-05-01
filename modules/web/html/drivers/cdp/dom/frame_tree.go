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
