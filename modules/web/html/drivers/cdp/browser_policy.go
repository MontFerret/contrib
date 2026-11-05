package cdp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/logging"
)

const policyCleanupTimeout = 5 * time.Second

// browserPolicy owns one browser-wide Fetch session. Page sessions never disable
// it, and failure permanently prevents this driver from accepting more work.
type browserPolicy struct {
	mu          sync.Mutex
	filter      *network.Filter
	conn        *rpcc.Conn
	interceptor *network.Interceptor
	sessions    map[*policySession]struct{}
	err         error
	closed      bool
	closeDone   chan struct{}
	closeErr    error
	watchDone   chan struct{}
}

func newBrowserPolicy(origins []string) *browserPolicy {
	filter := &network.Filter{AllowedOrigins: append([]string{}, origins...)}
	_, err := network.NewInterceptorFilter(filter)

	return &browserPolicy{
		filter:    filter,
		err:       err,
		sessions:  make(map[*policySession]struct{}),
		closeDone: make(chan struct{}),
	}
}

func (p *browserPolicy) begin(ctx context.Context, driver *Driver) (*policySession, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, errors.New("CDP driver is closed")
	}

	if p.err != nil {
		return nil, p.err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if p.conn == nil {
		if err := p.initialize(ctx, driver); err != nil {
			return nil, err
		}
	}

	// Check synchronously as well as in the watcher, so a new operation cannot
	// slip through while the failure notification is waiting for this mutex.
	if p.conn.Context().Err() != nil || p.interceptor.Err() != nil || !p.interceptor.IsRunning() {
		return nil, errors.New("CDP browser policy connection is unavailable; create a new driver")
	}

	operationCtx, cancel := context.WithCancel(ctx)
	lifetime, invalidate := context.WithCancel(context.Background())
	session := &policySession{policy: p, ctx: operationCtx, cancel: cancel, lifetime: lifetime, invalidate: invalidate}
	p.sessions[session] = struct{}{}

	return session, nil
}

func (p *browserPolicy) initialize(ctx context.Context, driver *Driver) error {
	startupCtx, cancel := context.WithTimeout(ctx, policyCleanupTimeout)
	defer cancel()

	conn, client, err := driver.openBrowser(startupCtx)
	if err != nil {
		return fmt.Errorf("initialize CDP browser policy: %w", err)
	}

	interceptor := network.NewInterceptor(logging.From(ctx), client)
	if err := interceptor.AddFilter("allowed_origins", p.filter); err != nil {
		_ = conn.Close()

		return err
	}

	// Startup cancellation must interrupt Fetch.enable without making the
	// caller's operation context the lifetime of the persistent connection.
	policyCtx, stop := context.WithCancel(conn.Context())
	stopStartup := context.AfterFunc(startupCtx, stop)
	err = interceptor.Run(policyCtx)
	stopStartup()
	if startupCtx.Err() != nil {
		err = startupCtx.Err()
	}

	if err != nil {
		stop()
		_ = conn.Close()
		_ = interceptor.Close()

		return fmt.Errorf("enable CDP browser policy: %w", err)
	}

	p.conn = conn
	p.interceptor = interceptor
	p.watchDone = make(chan struct{})
	go p.watch(conn, interceptor)

	return nil
}

func (p *browserPolicy) watch(conn *rpcc.Conn, interceptor *network.Interceptor) {
	defer close(p.watchDone)

	select {
	case <-conn.Context().Done():
	case <-interceptor.Done():
	}

	err := interceptor.Err()
	if err == nil {
		err = errors.New("browser policy connection or interception stream closed")
	}

	p.mu.Lock()
	if p.closed || p.err != nil {
		p.mu.Unlock()

		return
	}

	p.err = fmt.Errorf("CDP browser policy lost; create a new driver: %w", err)
	sessions := p.snapshot()
	p.mu.Unlock()

	_ = p.closeSessions(sessions)
	_ = conn.Close()
	_ = interceptor.Close()
}

func (p *browserPolicy) release(session *policySession) {
	p.mu.Lock()
	delete(p.sessions, session)
	p.mu.Unlock()
}

// snapshot is called with mu held; cleanup must run after releasing it.
func (p *browserPolicy) snapshot() []*policySession {
	sessions := make([]*policySession, 0, len(p.sessions))
	for session := range p.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (p *browserPolicy) closeSessions(sessions []*policySession) error {
	results := make(chan error, len(sessions))
	for _, session := range sessions {
		go func() { results <- session.Close() }()
	}

	var errs []error
	for range sessions {
		errs = append(errs, <-results)
	}

	return errors.Join(errs...)
}

func (p *browserPolicy) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		<-p.closeDone

		return p.closeErr
	}

	p.closed = true
	sessions := p.snapshot()
	conn, interceptor, watchDone := p.conn, p.interceptor, p.watchDone
	p.mu.Unlock()

	err := p.closeSessions(sessions)
	if interceptor != nil {
		err = errors.Join(err, interceptor.Close())
	}

	if conn != nil {
		err = errors.Join(err, conn.Close())
	}

	if watchDone != nil {
		<-watchDone
	}

	p.closeErr = err
	close(p.closeDone)

	return err
}
