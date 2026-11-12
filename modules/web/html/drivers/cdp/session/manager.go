package session

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"
)

var (
	attachRootSessionClient    = attachRootClient
	attachRelatedSessionClient = attachRelatedClient
	enableRelatedSessionClient = enableAttachedClient
	syncEventStreams           = func(
		attached target.AttachedToTargetClient,
		detached target.DetachedFromTargetClient,
		message target.ReceivedMessageFromTargetClient,
	) error {
		return cdp.Sync(attached, detached, message)
	}
)

type Manager struct {
	targets          *targetTracker
	detached         target.DetachedFromTargetClient
	attached         target.AttachedToTargetClient
	message          target.ReceivedMessageFromTargetClient
	ctx              context.Context
	cancel           context.CancelFunc
	clientsBySession map[target.SessionID]*Client
	browserClient    *cdp.Client
	listeners        map[ListenerID]Listener
	browserConn      *rpcc.Conn
	clientsByTarget  map[target.ID]*Client
	rootTargetID     target.ID
	wg               sync.WaitGroup
	nextListenerID   atomic.Int64
	mu               sync.RWMutex
	closeOnce        sync.Once
	closeErr         error
}

func New(
	ctx context.Context,
	browserConn *rpcc.Conn,
	browserClient *cdp.Client,
	rootTargetID target.ID,
) (*Manager, error) {
	if browserClient == nil {
		return nil, errors.New("browser client is required")
	}

	managerCtx, cancel := context.WithCancel(context.Background())

	attached, err := browserClient.Target.AttachedToTarget(managerCtx)
	if err != nil {
		cancel()
		return nil, err
	}

	detached, err := browserClient.Target.DetachedFromTarget(managerCtx)
	if err != nil {
		cancel()
		_ = attached.Close()
		return nil, err
	}

	message, err := browserClient.Target.ReceivedMessageFromTarget(managerCtx)
	if err != nil {
		cancel()
		_ = attached.Close()
		_ = detached.Close()
		return nil, err
	}

	if err := syncEventStreams(attached, detached, message); err != nil {
		cancel()
		_ = attached.Close()
		_ = detached.Close()
		_ = message.Close()
		return nil, err
	}

	rootClient, err := attachRootSessionClient(ctx, browserClient, rootTargetID)
	if err != nil {
		cancel()
		_ = attached.Close()
		_ = detached.Close()
		_ = message.Close()
		return nil, err
	}

	manager := &Manager{
		browserConn:      browserConn,
		browserClient:    browserClient,
		rootTargetID:     rootTargetID,
		ctx:              managerCtx,
		cancel:           cancel,
		attached:         attached,
		detached:         detached,
		message:          message,
		clientsBySession: make(map[target.SessionID]*Client),
		clientsByTarget:  make(map[target.ID]*Client),
		listeners:        make(map[ListenerID]Listener),
	}
	manager.registerClient(rootClient)

	manager.wg.Add(2)
	go manager.routeMessages()
	go manager.watchTargets()

	if err := browserClient.Target.AutoAttachRelated(
		ctx,
		target.NewAutoAttachRelatedArgs(rootTargetID, false),
	); err != nil {
		return nil, errors.Join(err, manager.Close())
	}

	return manager, nil
}

// IsFlat reports whether the manager routes top-level CDP session identifiers.
func (m *Manager) IsFlat() bool { return m != nil && m.targets != nil }

// Context ends when the manager closes or its flat browser connection is lost.
func (m *Manager) Context() context.Context {
	if m == nil || m.ctx == nil {
		return context.Background()
	}

	return m.ctx
}

func (m *Manager) Root() *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.clientsByTarget[m.rootTargetID]
}

func (m *Manager) Snapshot() []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make([]*Client, 0, len(m.clientsBySession))
	for _, client := range m.clientsBySession {
		snapshot = append(snapshot, client)
	}

	return snapshot
}

func (m *Manager) ClientByTarget(targetID target.ID) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, ok := m.clientsByTarget[targetID]

	return client, ok
}

func (m *Manager) AddListener(listener Listener) ListenerID {
	if listener == nil {
		return 0
	}

	id := ListenerID(m.nextListenerID.Add(1))

	m.mu.Lock()
	m.listeners[id] = listener
	m.mu.Unlock()

	return id
}

func (m *Manager) RemoveListener(id ListenerID) {
	if id == 0 {
		return
	}

	m.mu.Lock()
	delete(m.listeners, id)
	m.mu.Unlock()
}

func (m *Manager) Close() error {
	m.closeOnce.Do(func() {
		if m.targets != nil {
			m.targets.endPage(rpcc.ErrConnClosing)
		}

		m.cancel()
		if m.targets != nil {
			m.targets.events.stop(context.Canceled)
			if m.browserConn.Context().Err() == nil {
				m.closeErr = m.browserConn.Close()
			}

			m.wg.Wait()

			return
		}

		clients := m.Snapshot()
		errs := make([]error, 0, len(clients)+4)

		for _, client := range clients {
			if m.browserConn != nil && m.browserConn.Context().Err() != nil {
				// The browser channel is already gone; a remote detach cannot run.
				client.markDetached()
			}

			if err := client.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		if m.attached != nil {
			if err := m.attached.Close(); err != nil && !isIgnorableManagerStreamCloseError(err) {
				errs = append(errs, err)
			}
		}

		if m.detached != nil {
			if err := m.detached.Close(); err != nil && !isIgnorableManagerStreamCloseError(err) {
				errs = append(errs, err)
			}
		}

		if m.message != nil {
			if err := m.message.Close(); err != nil && !isIgnorableManagerStreamCloseError(err) {
				errs = append(errs, err)
			}
		}

		if m.browserConn != nil && m.browserConn.Context().Err() == nil {
			if err := m.browserConn.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		m.wg.Wait()
		m.closeErr = errors.Join(errs...)
	})

	return m.closeErr
}

func (m *Manager) routeMessages() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.message.Ready():
			reply, err := m.message.Recv()
			if err != nil {
				if m.ctx.Err() != nil {
					return
				}

				return
			}

			client := m.clientBySession(reply.SessionID)
			if client == nil {
				continue
			}

			if err := client.writeMessage([]byte(reply.Message)); err != nil {
				m.unregisterClient(reply.SessionID)
			}
		}
	}
}

func (m *Manager) watchTargets() {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.attached.Ready():
			reply, err := m.attached.Recv()
			if err != nil {
				if m.ctx.Err() != nil {
					return
				}

				return
			}

			if err := m.handleAttached(reply); err != nil {
				continue
			}
		case <-m.detached.Ready():
			reply, err := m.detached.Recv()
			if err != nil {
				if m.ctx.Err() != nil {
					return
				}

				return
			}

			m.handleDetached(reply)
		}
	}
}

func (m *Manager) handleAttached(reply *target.AttachedToTargetReply) error {
	if reply == nil {
		return nil
	}

	if reply.TargetInfo.TargetID == m.rootTargetID {
		return nil
	}

	if existing := m.clientBySession(reply.SessionID); existing != nil {
		return nil
	}

	client, err := attachRelatedSessionClient(m.ctx, m.browserClient, reply)
	if err != nil {
		return err
	}

	if client == nil {
		return nil
	}

	m.registerClient(client)

	if err := enableRelatedSessionClient(m.ctx, client.CDP); err != nil {
		m.unregisterClient(client.ID)
		client.markDetached()
		_ = client.Close()

		return err
	}

	m.emit(Event{Kind: EventAttached, Client: client})

	return nil
}

func (m *Manager) handleDetached(reply *target.DetachedFromTargetReply) {
	if reply == nil {
		return
	}

	client := m.unregisterClient(reply.SessionID)
	if client == nil {
		return
	}

	client.markDetached()
	_ = client.Close()

	m.emit(Event{Kind: EventDetached, Client: client})
}

func (m *Manager) registerClient(client *Client) {
	if client == nil {
		return
	}

	m.mu.Lock()
	m.clientsBySession[client.ID] = client
	m.clientsByTarget[client.TargetID] = client
	m.mu.Unlock()
}

func (m *Manager) unregisterClient(sessionID target.SessionID) *Client {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clientsBySession[sessionID]
	if !ok {
		return nil
	}

	delete(m.clientsBySession, sessionID)
	delete(m.clientsByTarget, client.TargetID)

	return client
}

func (m *Manager) clientBySession(sessionID target.SessionID) *Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.clientsBySession[sessionID]
}

func (m *Manager) emit(event Event) {
	m.mu.RLock()

	listeners := make([]Listener, 0, len(m.listeners))
	for _, listener := range m.listeners {
		listeners = append(listeners, listener)
	}

	m.mu.RUnlock()

	for _, listener := range listeners {
		listener(event)
	}
}

// FrameSessions returns the active out-of-process frames in the root document
// hierarchy and any initialization error affecting frame enumeration.
func (m *Manager) FrameSessions() ([]*Client, error) {
	if m == nil || m.targets == nil {
		return nil, nil
	}

	frames, problems := m.FrameTopology()
	var errs []error
	for _, err := range problems {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return frames, errors.Join(errs...)
}

// FrameError returns the initialization failure of the specified root-page frame.
func (m *Manager) FrameError(id target.ID) error {
	if m == nil || m.targets == nil {
		return nil
	}

	m.targets.mu.RLock()
	defer m.targets.mu.RUnlock()

	return m.targets.framesFailed[id]
}

// PageContext ends when the root target closes. The manager continues to own surviving popups.
func (m *Manager) PageContext() context.Context {
	if m.IsFlat() {
		return m.targets.pageCtx
	}

	return m.Context()
}

// FrameTopology snapshots published renderers and unresolved root-hierarchy frame transitions.
// A nil problem denotes an in-progress renderer change.
func (m *Manager) FrameTopology() ([]*Client, map[target.ID]error) {
	if !m.IsFlat() {
		return nil, nil
	}

	t := m.targets
	t.mu.RLock()
	defer t.mu.RUnlock()
	m.mu.RLock()
	defer m.mu.RUnlock()
	var clients []*Client
	for _, c := range m.clientsBySession {
		if c.Info.Type == "iframe" && !c.popup {
			clients = append(clients, c)
		}
	}

	problems := make(map[target.ID]error, len(t.transitions)+len(t.framesFailed))
	for id := range t.transitions {
		problems[id] = nil
	}

	for id, err := range t.framesFailed {
		problems[id] = err
	}

	return clients, problems
}

// FrameChanging reports whether an existing document must resolve a replacement renderer.
func (m *Manager) FrameChanging(id target.ID) bool {
	if !m.IsFlat() {
		return false
	}

	t := m.targets
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, pending := t.transitions[id]

	return pending || t.framesFailed[id] != nil
}
