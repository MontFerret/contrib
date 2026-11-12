package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

type targetChange struct {
	frameID        target.ID
	frameReason    string
	source         target.SessionID
	session        *targetSession
	info           target.Info
	detached       target.SessionID
	destroyed      target.ID
	initialized    *Client
	failedPopup    target.ID
	discoveryReady bool
	cleanup        *Client
	waiting        bool
	err            error
}

type targetTracker struct {
	manager      *Manager
	router       *router
	events       *mailbox[targetChange]
	root         chan error
	pageCtx      context.Context
	endPage      context.CancelCauseFunc
	setupTimeout time.Duration
	// The event loop owns target selection and serializes listener notifications.
	active       map[target.ID]*Client
	bySession    map[target.SessionID]*Client
	mu           sync.RWMutex
	owned        map[target.ID]bool
	framesFailed map[target.ID]error
	transitions  map[target.ID]target.SessionID
	jobs         sync.WaitGroup
}

// NewFlat attaches to root on a dedicated browser connection. On success the
// manager owns the connection. On error the caller retains it for rollback.
func NewFlat(ctx context.Context, conn *FlatConnection, root target.ID) (*Manager, error) {
	if conn == nil || conn.Conn == nil || conn.router == nil || root == "" {
		return nil, errors.New("flat browser connection and root target are required")
	}

	lifetime, cancel := context.WithCancel(context.Background())
	m := &Manager{
		ctx:              lifetime,
		cancel:           cancel,
		browserConn:      conn.Conn,
		browserClient:    cdp.NewClient(conn.Conn),
		rootTargetID:     root,
		clientsByTarget:  make(map[target.ID]*Client),
		clientsBySession: make(map[target.SessionID]*Client),
		listeners:        make(map[ListenerID]Listener),
	}

	pageCtx, endPage := context.WithCancelCause(lifetime)
	t := &targetTracker{
		pageCtx:      pageCtx,
		endPage:      endPage,
		setupTimeout: 30 * time.Second,
		transitions:  make(map[target.ID]target.SessionID),
		manager:      m,
		router:       conn.router,
		events:       newMailbox[targetChange](),
		root:         make(chan error, 1),
		active:       make(map[target.ID]*Client),
		bySession:    make(map[target.SessionID]*Client),
		owned:        map[target.ID]bool{root: true},
		framesFailed: make(map[target.ID]error),
	}

	m.targets = t
	conn.router.mu.Lock()

	if conn.router.targets != nil || conn.router.err != nil {
		err := errors.Join(errors.New("flat browser connection unavailable"), conn.router.err)
		conn.router.mu.Unlock()
		cancel()

		return nil, err
	}

	conn.router.targets = t
	conn.router.mu.Unlock()

	m.wg.Go(t.run)

	err := m.browserClient.Target.SetDiscoverTargets(ctx, target.NewSetDiscoverTargetsArgs(true))
	if err == nil {
		t.events.put(targetChange{discoveryReady: true})
		err = m.browserClient.Target.AutoAttachRelated(ctx, target.NewAutoAttachRelatedArgs(root, false))
	}

	if err == nil {
		select {
		case err = <-t.root:
		case <-ctx.Done():
			err = ctx.Err()
		case <-conn.Conn.Context().Done():
			err = rpcc.ErrConnClosing
		}

		if err == nil {
			err = ctx.Err()
		}

		if err == nil && m.Root() == nil {
			err = errors.New("root target detached during initialization")
		}
	}

	if err != nil {
		cancel()
		t.events.stop(err)
		m.wg.Wait()

		return nil, err
	}

	return m, nil
}

func (t *targetTracker) run() {
	defer func() {
		t.endPage(rpcc.ErrConnClosing)
		t.manager.cancel()
		t.router.mu.Lock()
		for _, s := range t.router.sessions {
			t.router.remove(s)
		}

		t.router.mu.Unlock()
		t.jobs.Wait()
		for _, c := range t.active {
			// The router has relinquished every session. Finish local rpcc
			// closure before publishing detach or returning from Manager.Close.
			_ = c.Conn.Close()

			if t.manager.unregisterClient(c.ID) != nil {
				t.manager.emit(Event{Kind: EventDetached, Client: c})
			}
		}

		clear(t.active)
		t.mu.Lock()
		clear(t.bySession)
		clear(t.framesFailed)
		t.mu.Unlock()
		t.events.clear()
	}()

	// Discovery can enumerate an existing nested popup before its opener.
	var initial []target.Info
	discovering := true
	for {
		ev, err := t.events.next()
		if err != nil || t.manager.ctx.Err() != nil {
			return
		}

		switch {
		case ev.frameID != "":
			t.frameChanged(ev)
		case ev.discoveryReady:
			t.finishDiscovery(initial)
			discovering, initial = false, nil
		case ev.cleanup != nil:
			t.manager.emit(Event{Kind: EventFailed, Client: ev.cleanup, Operation: "detach", Err: ev.err})
		case ev.detached != "":
			t.detach(ev.detached)
		case ev.destroyed != "":
			if ev.destroyed == t.manager.rootTargetID {
				t.endPage(rpcc.ErrConnClosing)
				select {
				case t.root <- errors.New("root target destroyed during initialization"):
				default:
				}
			}

			t.mu.Lock()
			delete(t.framesFailed, ev.destroyed)
			t.mu.Unlock()
		case ev.initialized != nil:
			t.publish(ev)
		case ev.session != nil && ev.err != nil:
			if c := t.bySession[ev.session.id]; c != nil {
				t.fail(c, ev.session, ev.err)
			}
		case ev.failedPopup != "":
			t.report(&Client{TargetID: ev.failedPopup, Info: ev.info, popup: true}, "attach popup", ev.err)
		case ev.session != nil:
			t.attach(ev)
		default:
			if !t.discover(ev.info) && discovering {
				initial = append(initial, ev.info)
			}
		}
	}
}

func (t *targetTracker) discover(info target.Info) bool {
	t.mu.Lock()

	if t.owned[info.TargetID] {
		t.mu.Unlock()

		return true
	}

	if info.OpenerID == nil || !t.owned[*info.OpenerID] {
		t.mu.Unlock()

		return false
	}

	t.owned[info.TargetID] = true
	t.mu.Unlock()

	t.jobs.Go(func() {
		_, err := t.manager.browserClient.Target.AttachToTarget(t.manager.ctx, target.NewAttachToTargetArgs(info.TargetID).SetFlatten(true))
		if err != nil {
			t.events.put(targetChange{failedPopup: info.TargetID, info: info, err: err})
		}
	})

	return true
}

func (t *targetTracker) attach(ev targetChange) {
	s, info := ev.session, ev.info
	root := info.TargetID == t.manager.rootTargetID
	popup := false

	if s.parent != "" {
		parent := t.bySession[s.parent]
		if parent == nil {
			t.discard(s, info)

			return
		}

		popup = parent.popup
	}

	switch info.Type {
	case "page":
		if !root {
			t.mu.RLock()
			popup = t.owned[info.TargetID] || (info.OpenerID != nil && t.owned[*info.OpenerID])
			t.mu.RUnlock()

			if !popup {
				t.discard(s, info)

				return
			}

			t.mu.Lock()
			t.owned[info.TargetID] = true
			t.mu.Unlock()
		}
	case "iframe":
		t.mu.Lock()
		t.owned[info.TargetID] = true
		t.mu.Unlock()
	case "worker", "service_worker", "shared_worker":
	default:
		t.discard(s, info)

		return
	}

	if current := t.active[info.TargetID]; current != nil {
		t.discard(s, info)

		return
	}

	client, err := s.client(t.manager.ctx, info, popup)
	if err != nil {
		t.report(&Client{ID: s.id, TargetID: info.TargetID, Info: info, popup: popup}, "open session", err)
		t.discard(s, info)

		return
	}

	t.mu.Lock()
	t.bySession[s.id] = client
	if info.Type == "iframe" && !popup {
		t.transitions[info.TargetID] = s.id
		delete(t.framesFailed, info.TargetID)
	}

	t.mu.Unlock()
	t.active[info.TargetID] = client

	if root {
		t.publish(targetChange{initialized: client, session: s})

		return
	}

	t.jobs.Go(func() {
		err := t.initialize(s.setup, client, ev.waiting)
		t.events.put(targetChange{initialized: client, session: s, err: err})
	})
}

func (t *targetTracker) initialize(ctx context.Context, c *Client, waiting bool) error {
	switch c.Info.Type {
	case "page", "iframe":
		if err := enableAttachedClient(ctx, c.CDP); err != nil {
			return err
		}

		// Descendants of a remote frame or popup attach on its own session.
		if err := c.CDP.Target.SetAutoAttach(ctx, target.NewSetAutoAttachArgs(true, false).SetFlatten(true)); err != nil {
			return err
		}
	default:
		if err := c.CDP.Runtime.Enable(ctx); err != nil {
			return err
		}

		if err := c.CDP.Network.Enable(ctx, network.NewEnableArgs()); err != nil {
			var protocol *rpcc.ResponseError

			if !errors.As(err, &protocol) || protocol.Code != -32601 {
				return err
			}

			c.runtimeOnly = true
		}
	}

	if waiting {
		return c.CDP.Runtime.RunIfWaitingForDebugger(ctx)
	}

	return nil
}

func (t *targetTracker) publish(ev targetChange) {
	c := ev.initialized
	if t.active[c.TargetID] != c {
		return
	}

	if ev.session.setup != nil {
		if cause := context.Cause(ev.session.setup); cause != nil {
			// The router queued the corresponding failure or detach under its lock.
			return
		}
	}

	if ev.err != nil {
		t.fail(c, ev.session, ev.err)

		return
	}

	t.mu.Lock()
	delete(t.framesFailed, c.TargetID)
	delete(t.transitions, c.TargetID)
	t.manager.registerClient(c)
	t.mu.Unlock()
	t.manager.emit(Event{Kind: EventAttached, Client: c})
	if err := ev.session.releaseEvents(); err != nil {
		if !errors.Is(err, rpcc.ErrConnClosing) {
			t.fail(c, ev.session, err)
		}

		return
	}

	if c.TargetID == t.manager.rootTargetID {
		select {
		case t.root <- nil:
		default:
		}
	}
}

func (t *targetTracker) fail(c *Client, s *targetSession, err error) {
	t.report(c, "initialize", err)
	t.detach(c.ID)
	t.discard(s, c.Info)
}

func (t *targetTracker) report(c *Client, op string, err error) {
	if t.manager.ctx.Err() != nil {
		return
	}

	err = fmt.Errorf("%s CDP %s target %s (session %s): %w", op, c.Info.Type, c.TargetID, c.ID, err)

	if c.Info.Type == "iframe" && !c.popup {
		t.mu.Lock()
		t.framesFailed[c.TargetID] = err
		t.mu.Unlock()
	}

	if c.TargetID == t.manager.rootTargetID {
		select {
		case t.root <- err:
		default:
		}
	}

	t.manager.emit(Event{Kind: EventFailed, Client: c, Err: err, Operation: op})
}

func (t *targetTracker) detach(id target.SessionID) {
	t.mu.Lock()
	c := t.bySession[id]
	delete(t.bySession, id)
	if c != nil && t.transitions[c.TargetID] == id {
		delete(t.transitions, c.TargetID)
	}

	t.mu.Unlock()

	if c == nil {
		return
	}

	delete(t.active, c.TargetID)
	if c.TargetID == t.manager.rootTargetID {
		t.endPage(rpcc.ErrConnClosing)
	}

	if t.manager.unregisterClient(id) != nil {
		t.manager.emit(Event{Kind: EventDetached, Client: c})
	}
}

func (t *targetTracker) parentClient(id target.SessionID) *cdp.Client {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if c := t.bySession[id]; c != nil {
		return c.CDP
	}

	return nil
}

// A remote detach is a command and cannot run on the lifecycle event loop.
func (t *targetTracker) discard(s *targetSession, info target.Info) {
	t.jobs.Go(func() {
		if err := s.detach(); err != nil {
			client := &Client{ID: s.id, TargetID: info.TargetID, Info: info}
			t.events.put(targetChange{cleanup: client, err: fmt.Errorf("detach CDP target %s (session %s): %w", info.TargetID, s.id, err)})
		}
	})
}

func (t *targetTracker) livePopups(infos []target.Info) []target.ID {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Chrome can return a grandchild before its opener. Keep historical
	// ownership because the opener may disappear before page cleanup.
	for changed := true; changed; {
		changed = false
		for _, info := range infos {
			if info.Type == "page" && !t.owned[info.TargetID] && info.OpenerID != nil && t.owned[*info.OpenerID] {
				t.owned[info.TargetID] = true
				changed = true
			}
		}
	}

	var ids []target.ID
	for _, info := range infos {
		if info.Type == "page" && info.TargetID != t.manager.rootTargetID && t.owned[info.TargetID] {
			ids = append(ids, info.TargetID)
		}
	}

	return ids
}

func (t *targetTracker) finishDiscovery(initial []target.Info) {
	for len(initial) > 0 {
		remaining := initial[:0]
		for _, info := range initial {
			if !t.discover(info) {
				remaining = append(remaining, info)
			}
		}

		if len(remaining) == len(initial) {
			break
		}

		initial = remaining
	}
}

func (t *targetTracker) frameChanged(ev targetChange) {
	t.mu.Lock()
	defer t.mu.Unlock()

	source := t.bySession[ev.source]
	if source == nil || source.popup {
		return
	}

	switch ev.frameReason {
	case "swap":
		if c, ok := t.manager.ClientByTarget(ev.frameID); !ok || c.ID == ev.source {
			t.transitions[ev.frameID] = ev.source
		}
	case "remove", "navigated":
		if ev.frameReason == "navigated" {
			// Page.enable can emit frameNavigated before setup completes.
			if c := t.bySession[t.transitions[ev.frameID]]; c != nil && c.TargetID == ev.frameID {
				if _, ready := t.manager.ClientByTarget(ev.frameID); !ready {
					return
				}
			}
		}

		delete(t.transitions, ev.frameID)
		if ev.frameReason == "remove" {
			delete(t.framesFailed, ev.frameID)
		}
	}
}
