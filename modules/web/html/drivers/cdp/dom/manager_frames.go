package dom

import (
	"context"

	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) SetMainFrame(doc *HTMLDocument) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.installMainFrame(doc, doc.Frame())
}

func (m *Manager) AddFrame(frame page.FrameTree) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.addFrameInternal(frame)
}

func (m *Manager) RemoveFrame(frameID page.FrameID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.removeFrameInternal(frameID)
}

func (m *Manager) RemoveFrameRecursively(frameID page.FrameID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.removeFrameRecursivelyInternal(frameID)
}

func (m *Manager) RemoveFramesByParentID(parentFrameID page.FrameID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	frame, found := m.frames.Get(parentFrameID)
	if !found {
		return runtime.Error(runtime.ErrNotFound, "parent frame")
	}

	for _, child := range frame.tree.ChildFrames {
		if err := m.removeFrameRecursivelyInternal(child.Frame.ID); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) GetFrameTree(_ context.Context, frameID page.FrameID) (page.FrameTree, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	frame, found := m.frames.Get(frameID)
	if !found {
		return page.FrameTree{}, runtime.ErrNotFound
	}

	return frame.tree, nil
}

func (m *Manager) GetFrameNodes(ctx context.Context) (runtime.List, error) {
	// Loading a frame can change index membership during traversal.
	m.mu.Lock()
	defer m.mu.Unlock()

	count := m.frames.Length()
	arr := runtime.NewArray(count)
	visited := make(map[page.FrameID]bool, count)

	var visit func(page.FrameID) error
	visit = func(id page.FrameID) error {
		if visited[id] {
			return nil
		}

		frame, found := m.frames.Get(id)
		if !found {
			return nil
		}

		loaded, err := m.loadFrame(ctx, frame)
		if err != nil {
			return err
		}

		visited[id] = true

		if err := arr.Append(ctx, loaded.node); err != nil {
			return err
		}

		// Lazy loading can replace the children of this frame.
		for _, child := range loaded.tree.ChildFrames {
			if err := visit(child.Frame.ID); err != nil {
				return err
			}
		}

		return nil
	}

	if main := m.mainFrame.Get(); main != "" {
		if err := visit(main); err != nil {
			return nil, err
		}
	}

	// AddFrame can index independent trees before a main frame is installed.
	if len(visited) != m.frames.Length() {
		frames := m.frames.ToSlice()
		children := make(map[page.FrameID]bool)
		for _, frame := range frames {
			for _, child := range frame.tree.ChildFrames {
				children[child.Frame.ID] = true
			}
		}

		for _, frame := range frames {
			id := frame.tree.Frame.ID
			if !visited[id] && !children[id] {
				if err := visit(id); err != nil {
					return nil, err
				}
			}
		}
	}

	return arr, nil
}

func (m *Manager) addFrameInternal(frame page.FrameTree) {
	m.frames.Set(frame.Frame.ID, Frame{
		tree: frame,
		node: nil,
	})

	for _, child := range frame.ChildFrames {
		m.addFrameInternal(child)
	}
}

func (m *Manager) addPreloadedFrame(doc *HTMLDocument) {
	frame := doc.Frame()
	m.frames.Set(frame.Frame.ID, Frame{
		tree: frame,
		node: doc,
	})

	for _, child := range frame.ChildFrames {
		m.addFrameInternal(child)
	}
}

func (m *Manager) installMainFrame(doc *HTMLDocument, tree page.FrameTree) {
	previous := m.frames.ToSlice()
	retained := collectFrameIDs(tree)

	for _, frame := range previous {
		if frame.node == nil || frame.node == doc {
			continue
		}
		if _, ok := retained[frame.tree.Frame.ID]; !ok {
			frame.node.detach()
		}
	}

	m.frames.Clear()
	m.mainFrame.Set(tree.Frame.ID)
	m.addFrameTreePreservingNodes(tree, doc, previous)
}

func (m *Manager) addFrameTreePreservingNodes(tree page.FrameTree, root *HTMLDocument, previous []Frame) {
	node := root
	if tree.Frame.ID != root.identity {
		node = nil
		for _, frame := range previous {
			if frame.tree.Frame.ID == tree.Frame.ID {
				node = frame.node
				break
			}
		}
	}

	m.frames.Set(tree.Frame.ID, Frame{tree: tree, node: node})
	for _, child := range tree.ChildFrames {
		m.addFrameTreePreservingNodes(child, root, previous)
	}
}

func (m *Manager) removeFrameInternal(frameID page.FrameID) error {
	current, exists := m.frames.Get(frameID)
	if !exists {
		return runtime.Error(runtime.ErrNotFound, "frame")
	}

	m.frames.Remove(frameID)
	m.owners.Remove(frameID)

	mainFrameID := m.mainFrame.Get()
	if frameID == mainFrameID {
		m.mainFrame.Reset()
	}

	if current.node == nil {
		return nil
	}

	return current.node.Close()
}

func (m *Manager) removeFrameRecursivelyInternal(frameID page.FrameID) error {
	parent, exists := m.frames.Get(frameID)
	if !exists {
		return runtime.Error(runtime.ErrNotFound, "frame")
	}

	for _, child := range parent.tree.ChildFrames {
		if err := m.removeFrameRecursivelyInternal(child.Frame.ID); err != nil {
			return err
		}
	}

	return m.removeFrameInternal(frameID)
}
