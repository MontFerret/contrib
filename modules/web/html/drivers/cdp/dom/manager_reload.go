package dom

import (
	"context"
	"errors"

	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/target"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func (m *Manager) ReloadRoot(ctx context.Context) error {
	ftRepl, err := m.frameTree(ctx, false, "")
	if err != nil {
		return err
	}

	if !m.sessions.IsFlat() {
		m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
		m.owners.Retain(collectFrameIDs(ftRepl.FrameTree))
	}

	doc := m.GetMainFrame()
	if doc != nil {
		currentLoader := doc.Frame().Frame.LoaderID
		nextLoader := ftRepl.FrameTree.Frame.LoaderID
		if currentLoader != "" && currentLoader == nextLoader {
			m.refreshFrameTree(doc, ftRepl.FrameTree)
			return nil
		}
	}

	var state *documentState

	if doc == nil {
		doc, err = m.LoadDocument(ctx, ftRepl.FrameTree)
		if err != nil {
			return err
		}

		state = doc.currentState()
	} else {
		state, err = doc.reload(ctx, ftRepl.FrameTree)
		if err != nil {
			return err
		}

		if state == nil {
			return nil
		}
	}

	tree, err := m.reloadRetainedDocuments(ctx, doc.Frame())
	if err != nil {
		return err
	}

	m.installMainFrameIfCurrent(doc, state, tree)

	return nil
}

func (m *Manager) RefreshFrames(ctx context.Context) error {
	ftRepl, err := m.frameTree(ctx, true, "")
	if err != nil {
		return err
	}

	doc := m.GetMainFrame()
	if doc == nil {
		return m.ReloadRoot(ctx)
	}

	if !m.sessions.IsFlat() {
		m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
		m.owners.Retain(collectFrameIDs(ftRepl.FrameTree))
	}

	m.refreshFrameTree(doc, ftRepl.FrameTree)

	return nil
}

func (m *Manager) refreshFrameTree(doc *HTMLDocument, tree page.FrameTree) {
	doc.updateFrameTree(tree)

	for _, current := range m.frames.ToSlice() {
		if current.node == nil || current.node == doc {
			continue
		}

		if frame, found := findFrame(tree, current.node.identity); found {
			current.node.updateFrameTree(frame)
		}
	}

	m.mu.Lock()
	m.installMainFrame(doc, tree)
	m.mu.Unlock()
}

// Reload parents before descendants, carrying each recovered subtree into the
// enclosing snapshot. Copy child slices so published generations remain untouched.
func (m *Manager) reloadRetainedDocuments(ctx context.Context, tree page.FrameTree) (page.FrameTree, error) {
	if len(tree.ChildFrames) == 0 {
		return tree, nil
	}

	children := make([]page.FrameTree, len(tree.ChildFrames))

	for i, child := range tree.ChildFrames {
		var state *documentState

		current, found := m.frames.Get(child.Frame.ID)
		if view := m.renderers.Load(); view != nil {
			if _, unavailable := view.unavailable[child.Frame.ID]; unavailable {
				children[i] = child

				continue
			}
		}

		if found && current.node != nil {
			loaded, err := current.node.reload(ctx, child)
			if err != nil {
				return page.FrameTree{}, err
			}

			if loaded == nil {
				children[i] = current.node.Frame()

				continue
			}

			state = loaded
			child = current.node.Frame()
		}

		updated, err := m.reloadRetainedDocuments(ctx, child)
		if err != nil {
			return page.FrameTree{}, err
		}

		if state != nil && !current.node.updateFrameTreeIfCurrent(state, updated) {
			updated = current.node.Frame()
		}

		children[i] = updated
	}

	tree.ChildFrames = children

	return tree, nil
}

// Keep the root generation stable while publishing its recovered frame tree.
func (m *Manager) installMainFrameIfCurrent(doc *HTMLDocument, state *documentState, tree page.FrameTree) {
	m.mu.Lock()
	defer m.mu.Unlock()

	doc.mu.Lock()
	defer doc.mu.Unlock()

	if state == nil || doc.state != state {
		return
	}

	state.frameTree = tree
	doc.fallback.frameTree = tree
	m.installMainFrame(doc, tree)
}

func (m *Manager) refreshDocument(ctx context.Context, doc *HTMLDocument, observedState *documentState) error {
	if err := m.sessions.FrameError(target.ID(doc.identity)); err != nil {
		return err
	}

	ftRepl, err := m.frameTree(ctx, false, doc.identity)
	if err != nil {
		return err
	}

	if !m.sessions.IsFlat() {
		m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
		m.owners.Retain(collectFrameIDs(ftRepl.FrameTree))
	}

	frame, found := findFrame(ftRepl.FrameTree, doc.identity)
	if !found {
		if !doc.detachIfCurrent(observedState) {
			return nil
		}

		return drivers.ErrDetached
	}

	state, _, err := m.loadDocumentState(ctx, doc, frame)
	if err != nil {
		if errors.Is(err, drivers.ErrDetached) && !doc.detachIfCurrent(observedState) {
			return nil
		}

		return err
	}

	if !doc.replaceStateIfCurrent(observedState, state) {
		return nil
	}

	if ftRepl.FrameTree.Frame.ID == doc.identity {
		tree, err := m.reloadRetainedDocuments(ctx, doc.Frame())
		if err != nil {
			return err
		}

		m.installMainFrameIfCurrent(doc, state, tree)
	} else {
		m.mu.Lock()
		doc.mu.RLock()

		if doc.state == state {
			m.frames.Set(doc.identity, Frame{tree: state.frameTree, node: doc})
		}

		doc.mu.RUnlock()
		m.mu.Unlock()
	}

	return nil
}
