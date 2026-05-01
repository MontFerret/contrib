package dom

import "context"

func (m *Manager) ReloadRoot(ctx context.Context) error {
	ftRepl, err := m.rootClient.Page.GetFrameTree(ctx)
	if err != nil {
		return err
	}

	ids := collectFrameIDs(ftRepl.FrameTree)
	m.owners.Set(ftRepl.FrameTree.Frame.ID, m.rootClient)
	m.owners.Retain(ids)

	doc, err := m.LoadDocument(ctx, ftRepl.FrameTree)
	if err != nil {
		return err
	}

	m.SetMainFrame(doc)

	return nil
}
