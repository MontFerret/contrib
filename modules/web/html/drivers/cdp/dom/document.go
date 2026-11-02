package dom

import (
	"context"
	"errors"
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
	fallback  documentFallback
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
	state := &documentState{
		client:     client,
		input:      inputs,
		eval:       exec,
		frameTree:  frames,
		generation: 1,
	}

	if rootElement != nil {
		rootElement.bindDocument(doc, state)
	}

	state.element = rootElement
	doc.replaceState(state)

	return doc
}

func newHTMLDocument(logger zerolog.Logger, manager *Manager, frame page.FrameTree) *HTMLDocument {
	return &HTMLDocument{
		logger: logutil.WithComponent(logger.With(), "html_document").Logger(),
		dom:    manager,
		fallback: documentFallback{
			frameTree: frame,
		},
		identity: frame.Frame.ID,
	}
}

func (doc *HTMLDocument) Close() error {
	doc.detach()

	return nil
}

func (doc *HTMLDocument) Frame() page.FrameTree {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.state != nil {
		return doc.state.frameTree
	}

	return doc.fallback.frameTree
}

func (doc *HTMLDocument) Eval() *eval.Runtime {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.state == nil {
		return nil
	}

	return doc.state.eval
}

func (doc *HTMLDocument) snapshot() (*documentState, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.state == nil {
		return nil, drivers.ErrDetached
	}

	return doc.state, nil
}

func (doc *HTMLDocument) snapshotFrame() (page.FrameTree, error) {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.state == nil {
		return page.FrameTree{}, drivers.ErrDetached
	}

	return doc.state.frameTree, nil
}

func (doc *HTMLDocument) currentState() *documentState {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	return doc.state
}

func (doc *HTMLDocument) currentElement() *HTMLElement {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.state != nil {
		return doc.state.element
	}

	return doc.fallback.element
}

func (doc *HTMLDocument) isCurrentState(state *documentState) bool {
	doc.mu.RLock()
	defer doc.mu.RUnlock()

	return state != nil && doc.state == state
}

func (doc *HTMLDocument) replaceState(state *documentState) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	doc.state = state
	doc.fallback.element = state.element
	doc.fallback.frameTree = state.frameTree
}

func (doc *HTMLDocument) replaceStateIfCurrent(observedState, state *documentState) bool {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if observedState == nil || doc.state != observedState {
		return false
	}

	doc.state = state
	doc.fallback.element = state.element
	doc.fallback.frameTree = state.frameTree

	return true
}

func (doc *HTMLDocument) updateFrameTree(frame page.FrameTree) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if doc.state == nil {
		return
	}

	doc.state.frameTree = frame
	doc.fallback.frameTree = frame
}

// updateFrameTreeIfCurrent prevents a recovered subtree from overwriting a newer generation.
func (doc *HTMLDocument) updateFrameTreeIfCurrent(state *documentState, frame page.FrameTree) bool {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if state == nil || doc.state != state {
		return false
	}

	state.frameTree = frame
	doc.fallback.frameTree = frame

	return true
}

func (doc *HTMLDocument) detach() {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if doc.state == nil {
		return
	}

	doc.fallback.element = doc.state.element
	doc.fallback.frameTree = doc.state.frameTree
	doc.state = nil
}

func (doc *HTMLDocument) detachIfCurrent(observedState *documentState) bool {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if observedState == nil || doc.state != observedState {
		return false
	}

	doc.fallback.element = doc.state.element
	doc.fallback.frameTree = doc.state.frameTree
	doc.state = nil

	return true
}

func (doc *HTMLDocument) refresh(ctx context.Context, observedState *documentState) error {
	doc.refreshMu.Lock()
	defer doc.refreshMu.Unlock()

	state, err := doc.snapshot()
	if err != nil {
		return err
	}

	if state != observedState {
		return nil
	}

	return doc.dom.refreshDocument(ctx, doc, observedState)
}

// reload returns the installed generation, or nil if another replacement won.
func (doc *HTMLDocument) reload(ctx context.Context, frame page.FrameTree) (*documentState, error) {
	doc.refreshMu.Lock()
	defer doc.refreshMu.Unlock()

	observedState, err := doc.snapshot()
	if err != nil {
		return nil, err
	}

	state, err := doc.dom.loadDocumentState(ctx, doc, frame)
	if err != nil {
		if errors.Is(err, drivers.ErrDetached) && !doc.detachIfCurrent(observedState) {
			return nil, nil
		}

		return nil, err
	}

	if !doc.replaceStateIfCurrent(observedState, state) {
		return nil, nil
	}

	return state, nil
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
