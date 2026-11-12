package dom

import (
	"context"
	"errors"
	"reflect"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

type HTMLDocument struct {
	logger     zerolog.Logger
	dom        *Manager
	state      *documentState
	identity   page.FrameID
	fallback   documentFallback
	mu         sync.RWMutex
	refreshing chan struct{}
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
		refreshing: make(chan struct{}, 1),
		logger:     logutil.WithComponent(logger.With(), "html_document").Logger(),
		dom:        manager,
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

// A concurrent refresh can publish document state before it acquires Manager.mu
// to update the index. Only patch the snapshot that the index still describes.
func (doc *HTMLDocument) reconcileFrameTree(indexed, updated page.FrameTree) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	if doc.state != nil && reflect.DeepEqual(doc.state.frameTree, indexed) {
		doc.state.frameTree = updated
		doc.fallback.frameTree = updated
	}
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
	select {
	case <-ctx.Done():
		return ctx.Err()
	case doc.refreshing <- struct{}{}:
	}

	defer func() { <-doc.refreshing }()

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
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case doc.refreshing <- struct{}{}:
	}

	defer func() { <-doc.refreshing }()

	observedState, err := doc.snapshot()
	if err != nil {
		return nil, err
	}

	state, _, err := doc.dom.loadDocumentState(ctx, doc, frame)
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

// Rebind outside the frame index lock. Refresh acquires both the document
// generation lock and the frame index when publishing its replacement.
func (doc *HTMLDocument) rendererChanged(state *documentState) bool {
	if doc.dom == nil || !doc.dom.sessions.IsFlat() {
		return false
	}

	if view := doc.dom.renderers.Load(); view != nil {
		if _, unresolved := view.unavailable[doc.identity]; unresolved {
			return true
		}
	}

	client, _ := doc.dom.frameRenderer(doc.identity)

	return (state.renderer != nil && state.renderer.Err() != nil) || doc.dom.sessions.FrameChanging(target.ID(doc.identity)) || client != state.client
}
