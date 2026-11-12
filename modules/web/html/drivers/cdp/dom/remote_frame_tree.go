package dom

import "github.com/mafredri/cdp/protocol/page"

// Remote roots may arrive before their parents. Graft by frame ID while
// preserving siblings already present in Chromium's frame tree.
func mergeRemoteFrames(root *page.FrameTree, remote []page.FrameTree) {
	byID := make(map[page.FrameID]page.FrameTree, len(remote))
	children := make(map[page.FrameID][]page.FrameID)
	for _, tree := range remote {
		id := tree.Frame.ID
		byID[id] = tree

		if parent := tree.Frame.ParentID; parent != nil {
			children[*parent] = append(children[*parent], id)
		}
	}

	var visit func(*page.FrameTree)
	visit = func(node *page.FrameTree) {
		id := node.Frame.ID

		if replacement, ok := byID[id]; ok {
			delete(byID, id)
			*node = replacement
		}

		for i := range node.ChildFrames {
			visit(&node.ChildFrames[i])
		}

		for _, childID := range children[id] {
			if child, ok := byID[childID]; ok {
				visit(&child)
				node.ChildFrames = append(node.ChildFrames, child)
			}
		}
	}

	visit(root)
}
