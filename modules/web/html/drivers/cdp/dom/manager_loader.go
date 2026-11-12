package dom

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	cdpruntime "github.com/mafredri/cdp/protocol/runtime"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/internal/recovery"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

func (m *Manager) LoadRootDocument(ctx context.Context) (*HTMLDocument, error) {
	ftRepl, err := m.frameTree(ctx, false, "")
	if err != nil {
		return nil, err
	}

	m.RecordFrameClient(ftRepl.FrameTree.Frame.ID, m.rootClient)

	return m.LoadDocument(ctx, ftRepl.FrameTree)
}

func (m *Manager) LoadDocument(ctx context.Context, frame page.FrameTree) (*HTMLDocument, error) {
	doc, _, err := m.loadDocument(ctx, frame)

	return doc, err
}

func (m *Manager) loadDocument(ctx context.Context, frame page.FrameTree) (*HTMLDocument, bool, error) {
	doc := newHTMLDocument(m.logger, m, frame)
	state, recovered, err := m.loadDocumentState(ctx, doc, frame)
	if err != nil {
		return nil, false, err
	}

	doc.replaceState(state)

	return doc, recovered, nil
}

func (m *Manager) loadDocumentState(ctx context.Context, doc *HTMLDocument, frame page.FrameTree) (*documentState, bool, error) {
	var client *cdp.Client
	var renderer context.Context
	var exec *eval.Runtime
	var ref cdpruntime.RemoteObject
	refreshFrame := false

	_, err := recovery.Wait(ctx, func(ctx context.Context) (runtime.Value, error) {
		if refreshFrame {
			reply, err := m.frameTree(ctx, false, doc.identity)
			if err != nil {
				return runtime.None, err
			}

			current, found := findFrame(reply.FrameTree, doc.identity)
			if !found {
				return runtime.None, drivers.ErrDetached
			}

			frame = current
			m.RecordFrameClient(reply.FrameTree.Frame.ID, m.rootClient)
		}

		if ctx.Err() != nil {
			return runtime.None, ctx.Err()
		}

		client, renderer = m.frameRenderer(frame.Frame.ID)

		created, err := eval.Create(ctx, m.logger, client, frame.Frame.ID)
		if err != nil {
			return runtime.None, err
		}

		exec = created

		if ctx.Err() != nil {
			return runtime.None, ctx.Err()
		}

		root, err := exec.EvalRef(ctx, templates.GetDocument())
		if err != nil {
			return runtime.None, runtime.Error(err, "failed to load root element")
		}

		ref = root

		return runtime.True, nil
	}, func() {
		exec = nil
		refreshFrame = true
	}, 1)
	if err != nil {
		return nil, false, err
	}

	mouse := m.mouse
	if renderer != nil {
		value, loaded := m.rendererMice.LoadOrStore(client, input.NewMouse(client))

		mouse = value.(*input.Mouse)
		if !loaded {
			context.AfterFunc(renderer, func() { m.rendererMice.Delete(client) })
		}
	}

	inputs := input.New(m.logger, client, exec, m.keyboard, mouse)
	exec.SetLoader(NewNodeLoader(m))

	generation := uint64(1)
	if current := doc.currentState(); current != nil {
		generation = current.generation + 1
	}

	state := &documentState{
		renderer:   renderer,
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

	return state, refreshFrame, nil
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
	m.mu.Lock()
	defer m.mu.Unlock()

	frame, found := m.frames.Get(frameID)
	if !found {
		return nil, runtime.ErrNotFound
	}

	loaded, err := m.loadFrame(ctx, frame)

	return loaded.node, err
}

// The caller holds m.mu through lazy acquisition and installation.
func (m *Manager) loadFrame(ctx context.Context, frame Frame) (Frame, error) {
	if frame.node != nil {
		return frame, nil
	}

	doc, recovered, err := m.loadDocument(ctx, frame.tree)
	if err != nil {
		return Frame{}, runtime.Error(err, "failed to load frame document")
	}

	loaded := Frame{tree: doc.Frame(), node: doc}
	if recovered {
		if err := m.installRecoveredFrame(loaded, frame.tree); err != nil {
			return Frame{}, err
		}
	} else {
		m.frames.Set(loaded.tree.Frame.ID, loaded)
	}

	return loaded, nil
}

func (m *Manager) frameRenderer(id page.FrameID) (*cdp.Client, context.Context) {
	if view := m.renderers.Load(); view != nil {
		if owner := view.owners[id]; owner != nil {
			return owner.CDP, owner.Conn.Context()
		}
	} else if c, ok := m.owners.Get(id); ok && c != nil {
		return c, nil
	}

	return m.rootClient, nil
}
