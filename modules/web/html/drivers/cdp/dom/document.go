package dom

import (
	"context"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

type HTMLDocument struct {
	logger    zerolog.Logger
	dom       *Manager
	state     *documentState
	identity  page.FrameID
	mu        sync.RWMutex
	refreshMu sync.Mutex
}

func NewHTMLDocument(
	logger zerolog.Logger,
	client *cdp.Client,
	domManager *Manager,
	inputs *input.Manager,
	exec *eval.Runtime,
	rootElement *HTMLElement,
	frames page.FrameTree,
) *HTMLDocument {
	doc := newHTMLDocument(logger, domManager, frames)
	if rootElement != nil {
		rootElement.bindDocument(doc, exec, 1)
	}
	doc.replaceState(&documentState{
		client:     client,
		input:      inputs,
		eval:       exec,
		element:    rootElement,
		frameTree:  frames,
		generation: 1,
	})

	return doc
}

func newHTMLDocument(logger zerolog.Logger, manager *Manager, frame page.FrameTree) *HTMLDocument {
	return &HTMLDocument{
		logger:   logutil.WithComponent(logger.With(), "html_document").Logger(),
		dom:      manager,
		identity: frame.Frame.ID,
	}
}

func (doc *HTMLDocument) Close() error {
	doc.detach()

	return nil
}

func (doc *HTMLDocument) Frame() page.FrameTree {
	return doc.currentState().frameTree
}

func (doc *HTMLDocument) Eval() *eval.Runtime {
	return doc.currentState().eval
}

func (doc *HTMLDocument) snapshot() (*documentState, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.state == nil || !doc.state.active {
		return nil, drivers.ErrDetached
	}

	return doc.state, nil
}

func (doc *HTMLDocument) currentState() *documentState {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	return doc.state
}

func (doc *HTMLDocument) isGenerationCurrent(generation uint64) bool {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	return doc.state != nil && doc.state.active && doc.state.generation == generation
}

func (doc *HTMLDocument) generationChanged(generation uint64) bool {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	return doc.state != nil && doc.state.active && doc.state.generation != generation
}

func (doc *HTMLDocument) replaceState(state *documentState) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if doc.state != nil {
		doc.state.active = false
	}

	state.active = true
	doc.state = state
}

func (doc *HTMLDocument) updateFrameTree(frame page.FrameTree) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if doc.state == nil {
		return
	}

	state := *doc.state
	state.frameTree = frame
	doc.state = &state
}

func (doc *HTMLDocument) detach() {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if doc.state != nil {
		doc.state.active = false
	}
}

func (doc *HTMLDocument) refresh(ctx context.Context, observedGeneration uint64) error {
	doc.refreshMu.Lock()
	defer doc.refreshMu.Unlock()

	state, err := doc.snapshot()
	if err != nil {
		return err
	}

	if state.generation != observedGeneration {
		return nil
	}

	return doc.dom.refreshDocument(ctx, doc)
}

func (doc *HTMLDocument) reload(ctx context.Context, frame page.FrameTree) error {
	doc.refreshMu.Lock()
	defer doc.refreshMu.Unlock()

	state, err := doc.dom.loadDocumentState(ctx, doc, frame)
	if err != nil {
		return err
	}

	doc.replaceState(state)

	return nil
}

func (doc *HTMLDocument) logError(err error) *zerolog.Event {
	frame := doc.Frame().Frame

	return doc.logger.
		Error().
		Timestamp().
		Str("url", frame.URL).
		Str("securityOrigin", frame.SecurityOrigin).
		Str("mimeType", frame.MimeType).
		Str("frameID", string(frame.ID)).
		Err(err)
}
