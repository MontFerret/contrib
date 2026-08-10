package dom

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) LoadRootDocument(ctx context.Context) (*HTMLDocument, error) {
	ftRepl, err := m.rootClient.Page.GetFrameTree(ctx)
	if err != nil {
		return nil, err
	}

	m.RecordFrameClient(ftRepl.FrameTree.Frame.ID, m.rootClient)

	return m.LoadDocument(ctx, ftRepl.FrameTree)
}

func (m *Manager) LoadDocument(ctx context.Context, frame page.FrameTree) (*HTMLDocument, error) {
	doc := newHTMLDocument(m.logger, m, frame)
	state, err := m.loadDocumentState(ctx, doc, frame)
	if err != nil {
		return nil, err
	}

	doc.replaceState(state)

	return doc, nil
}

func (m *Manager) loadDocumentState(ctx context.Context, doc *HTMLDocument, frame page.FrameTree) (*documentState, error) {
	client := m.clientForFrame(frame.Frame.ID)

	exec, err := eval.Create(ctx, m.logger, client, frame.Frame.ID)
	if err != nil {
		return nil, err
	}

	inputs := input.New(m.logger, client, exec, m.keyboard, m.mouse)

	ref, err := exec.EvalRef(ctx, templates.GetDocument())
	if err != nil {
		return nil, runtime.Error(err, "failed to load root element")
	}

	exec.SetLoader(NewNodeLoader(m))

	generation := uint64(1)
	if current := doc.currentState(); current != nil {
		generation = current.generation + 1
	}

	state := &documentState{
		client:     client,
		input:      inputs,
		eval:       exec,
		frameTree:  frame,
		generation: generation,
	}
	state.element = newHTMLElement(
		m.logger,
		client,
		m,
		inputs,
		exec,
		*ref.ObjectID,
		doc,
		state,
	)

	return state, nil
}

func (m *Manager) ResolveElement(ctx context.Context, frameID page.FrameID, id cdpruntime.RemoteObjectID) (*HTMLElement, error) {
	doc, err := m.GetFrameNode(ctx, frameID)
	if err != nil {
		return nil, err
	}
	state, err := doc.snapshot()
	if err != nil {
		return nil, err
	}

	return newHTMLElement(
		m.logger,
		state.client,
		m,
		state.input,
		state.eval,
		id,
		doc,
		state,
	), nil
}

func (m *Manager) GetFrameNode(ctx context.Context, frameID page.FrameID) (*HTMLDocument, error) {
	return m.getFrameInternal(ctx, frameID)
}

func (m *Manager) getFrameInternal(ctx context.Context, frameID page.FrameID) (*HTMLDocument, error) {
	frame, found := m.frames.Get(frameID)
	if !found {
		return nil, runtime.ErrNotFound
	}

	// frame is initialized
	if frame.node != nil {
		return frame.node, nil
	}

	// the frame is not loaded yet
	doc, err := m.LoadDocument(ctx, frame.tree)
	if err != nil {
		return nil, runtime.Error(err, "failed to load frame document")
	}

	frame.node = doc
	m.frames.Set(frameID, frame)

	return doc, nil
}

func (m *Manager) clientForFrame(frameID page.FrameID) *cdp.Client {
	client, ok := m.owners.Get(frameID)
	if ok && client != nil {
		return client
	}

	return m.rootClient
}
