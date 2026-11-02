package dom

import "github.com/mafredri/cdp/protocol/page"

// installRecoveredFrame reconciles only the recovered branch. The caller holds m.mu.
func (m *Manager) installRecoveredFrame(recovered Frame, old page.FrameTree) error {
	tree := recovered.tree
	previous := m.frames.ToSlice()
	retained := collectFrameIDs(tree)

	for id := range collectFrameIDs(old) {
		if _, ok := retained[id]; !ok && m.frames.Has(id) {
			if err := m.removeFrameInternal(id); err != nil {
				return err
			}
		}
	}

	for _, frame := range previous {
		id := frame.tree.Frame.ID
		if _, ok := retained[id]; ok {
			if frame.node != nil {
				updated, _ := findFrame(tree, id)
				frame.node.reconcileFrameTree(frame.tree, updated)
			}

			continue
		}

		updated, found := replaceFrameSubtree(frame.tree, tree)
		if found {
			if frame.node != nil {
				frame.node.reconcileFrameTree(frame.tree, updated)
			}

			frame.tree = updated
			m.frames.Set(id, frame)
		}
	}

	m.addFrameTreePreservingNodes(tree, recovered.node, previous)

	return nil
}
