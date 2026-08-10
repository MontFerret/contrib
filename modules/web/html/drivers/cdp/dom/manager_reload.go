package dom

import (
	"context"

	"github.com/mafredri/cdp/protocol/page"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

func (m *Manager) ReloadRoot(ctx context.Context) error {
	ftRepl, err := m.rootClient.Page.GetFrameTree(ctx)
	if err != nil {
		return err
	}

	ids := collectFrameIDs(ftRepl.FrameTree)
	m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
	m.owners.Retain(ids)

	doc := m.GetMainFrame()
	if doc != nil {
		currentLoader := doc.Frame().Frame.LoaderID
		nextLoader := ftRepl.FrameTree.Frame.LoaderID
		if currentLoader != "" && currentLoader == nextLoader {
			m.refreshFrameTree(doc, ftRepl.FrameTree)
			return nil
		}
	}

	if doc == nil {
		doc, err = m.LoadDocument(ctx, ftRepl.FrameTree)
		if err != nil {
			return err
		}
	} else {
		replaced, err := doc.reload(ctx, ftRepl.FrameTree)
		if err != nil {
			return err
		}

		if !replaced {
			return nil
		}
	}

	if err := m.reloadRetainedDocuments(ctx, doc, ftRepl.FrameTree); err != nil {
		return err
	}

	m.mu.Lock()
	m.installMainFrame(doc, ftRepl.FrameTree)
	m.mu.Unlock()

	return nil
}

func (m *Manager) RefreshFrames(ctx context.Context) error {
	ftRepl, err := m.rootClient.Page.GetFrameTree(ctx)
	if err != nil {
		return err
	}

	doc := m.GetMainFrame()
	if doc == nil {
		return m.ReloadRoot(ctx)
	}

	ids := collectFrameIDs(ftRepl.FrameTree)
	m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
	m.owners.Retain(ids)
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

func (m *Manager) reloadRetainedDocuments(ctx context.Context, root *HTMLDocument, tree page.FrameTree) error {
	for _, current := range m.frames.ToSlice() {
		if current.node == nil || current.node == root {
			continue
		}

		frame, found := findFrame(tree, current.node.identity)
		if !found {
			continue
		}

		if _, err := current.node.reload(ctx, frame); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) refreshDocument(ctx context.Context, doc *HTMLDocument, observedState *documentState) error {
	ftRepl, err := m.rootClient.Page.GetFrameTree(ctx)
	if err != nil {
		return err
	}

	ids := collectFrameIDs(ftRepl.FrameTree)
	m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
	m.owners.Retain(ids)

	frame, found := findFrame(ftRepl.FrameTree, doc.identity)
	if !found {
		if !doc.detachIfCurrent(observedState) {
			return nil
		}

		return drivers.ErrDetached
	}

	state, err := m.loadDocumentState(ctx, doc, frame)
	if err != nil {
		return err
	}

	if !doc.replaceStateIfCurrent(observedState, state) {
		return nil
	}

	if ftRepl.FrameTree.Frame.ID == doc.identity {
		if err := m.reloadRetainedDocuments(ctx, doc, ftRepl.FrameTree); err != nil {
			return err
		}

		m.mu.Lock()
		m.installMainFrame(doc, ftRepl.FrameTree)
		m.mu.Unlock()
	} else {
		m.frames.Set(doc.identity, Frame{tree: frame, node: doc})
	}

	return nil
}
