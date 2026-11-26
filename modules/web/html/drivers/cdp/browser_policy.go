package cdp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/internal/cleanup"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	"github.com/MontFerret/ferret/v2/pkg/logging"
)

const policyCleanupTimeout = 5 * time.Second

type (
	policyInitialization struct {
		done   chan struct{}
		cancel context.CancelFunc
		err    error
		retry  bool
	}

	// browserPolicy owns one browser-wide Fetch session. Page sessions never disable
	// it, and failure permanently prevents this driver from accepting more work.
	browserPolicy struct {
		mu           sync.Mutex
		filter       *network.Filter
		conn         *rpcc.Conn
		interceptor  *network.Interceptor
		sessions     map[*policySession]struct{}
		err          error
		closed       bool
		cleanup      cleanup.State
		watchDone    chan struct{}
		initializing *policyInitialization
		cleanupErr   error
	}
)

func newBrowserPolicy(origins []string) *browserPolicy {
	filter := &network.Filter{AllowedOrigins: append([]string{}, origins...)}
	_, err := network.NewInterceptorFilter(filter)

	return &browserPolicy{
		filter:   filter,
		err:      err,
		sessions: make(map[*policySession]struct{}),
	}
}

func (p *browserPolicy) begin(ctx context.Context, driver *Driver) (*policySession, error) {
	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()

			return nil, errors.New("CDP driver is closed")
		}

		if p.err != nil {
			err := p.err
			p.mu.Unlock()

			return nil, err
		}

		if err := ctx.Err(); err != nil {
			p.mu.Unlock()

			return nil, err
		}

		if p.conn != nil {
			if p.conn.Context().Err() != nil || p.interceptor.Err() != nil || !p.interceptor.IsRunning() {
				p.mu.Unlock()

				return nil, errors.New("CDP browser policy connection is unavailable; create a new driver")
			}

			session := newPageSession(ctx, p, driver)
			p.sessions[session] = struct{}{}
			p.mu.Unlock()

			return session, nil
		}

		attempt := p.initializing
		if attempt != nil {
			p.mu.Unlock()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-attempt.done:
			}

			if err := ctx.Err(); err != nil {
				return nil, err
			}

			if attempt.err != nil && !attempt.retry {
				return nil, attempt.err
			}

			continue
		}

		startupCtx, cancel := context.WithTimeout(ctx, policyCleanupTimeout)
		attempt = &policyInitialization{done: make(chan struct{}), cancel: cancel}
		p.initializing = attempt
		p.mu.Unlock()

		p.initialize(startupCtx, ctx, driver, attempt)
		cancel()
		if attempt.err != nil {
			return nil, attempt.err
		}
	}
}

func (p *browserPolicy) initialize(ctx, caller context.Context, driver *Driver, attempt *policyInitialization) {
	var interceptor *network.Interceptor
	conn, client, err := driver.openBrowser(ctx)
	if err == nil {
		interceptor = network.NewInterceptor(logging.From(caller), client)
		err = interceptor.AddFilter("allowed_origins", p.filter)
		if err == nil {
			// Detach startup cancellation only after Fetch.enable has settled.
			policyCtx, stop := context.WithCancel(conn.Context())
			stopStartup := context.AfterFunc(ctx, stop)
			err = interceptor.Run(policyCtx)
			stopStartup()

			if ctx.Err() != nil {
				err = ctx.Err()
			}

			if err != nil {
				stop()
			}
		}
	}

	p.mu.Lock()
	if err == nil && !p.closed && ctx.Err() == nil {
		p.conn, p.interceptor = conn, interceptor
		p.watchDone = make(chan struct{})

		go p.watch(conn, interceptor)

		p.initializing = nil
		close(attempt.done)
		p.mu.Unlock()

		return
	}

	if err == nil {
		err = ctx.Err()
		if err == nil {
			err = errors.New("CDP driver is closed")
		}
	}

	p.mu.Unlock()

	// Do not make the next attempt available until every resource from this
	// attempt is released. Close can cancel startup without holding mu.
	if conn != nil {
		cleanupErr := conn.Close()
		p.recordCleanup(cleanupErr)
		err = errors.Join(err, cleanupErr)
	}

	if interceptor != nil {
		_ = interceptor.Close()
	}

	if contextErr := ctx.Err(); contextErr != nil && !errors.Is(err, contextErr) {
		err = errors.Join(err, contextErr)
	}

	p.mu.Lock()
	attempt.err = fmt.Errorf("initialize CDP browser policy: %w", err)
	attempt.retry = caller.Err() != nil && !p.closed
	p.initializing = nil
	close(attempt.done)
	p.mu.Unlock()
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

	ctx, cancel := defaultCleanupContext()
	defer cancel()

	for _, session := range sessions {
		session.invalidate()
		session.cancel()
	}

	// Managed cleanup belongs to the host, even after policy failure.
	p.closeSessions(ctx, sessions, true)
	p.recordCleanup(errors.Join(closeConnection(conn), interceptor.Close()))
}

func (p *browserPolicy) release(session *policySession, err error) {
	p.mu.Lock()
	if err != nil {
		p.cleanupErr = errors.Join(p.cleanupErr, err)
	}

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

func (p *browserPolicy) closeSessions(ctx context.Context, sessions []*policySession, unmanagedOnly bool) {
	var wg sync.WaitGroup
	for _, session := range sessions {
		if unmanagedOnly && session.managed() {
			continue
		}

		wg.Go(func() {
			_ = session.CloseContext(ctx)
			// Policy remains active until all owned page workers have settled.
			_ = session.CloseContext(context.Background())
		})
	}

	wg.Wait()
}

func (p *browserPolicy) recordCleanup(err error) {
	if err == nil {
		return
	}

	p.mu.Lock()
	p.cleanupErr = errors.Join(p.cleanupErr, err)
	p.mu.Unlock()
}

func (p *browserPolicy) Close() error {
	ctx, cancel := defaultCleanupContext()
	defer cancel()

	return p.CloseContext(ctx)
}

func (p *browserPolicy) CloseContext(ctx context.Context) error {
	p.mu.Lock()
	p.closed = true
	if p.initializing != nil {
		p.initializing.cancel()
	}

	p.mu.Unlock()

	return p.cleanup.Run(ctx, p.close)
}

func (p *browserPolicy) close(ctx context.Context) error {
	p.mu.Lock()
	sessions := p.snapshot()
	conn, interceptor, watchDone := p.conn, p.interceptor, p.watchDone
	attempt := p.initializing
	p.mu.Unlock()

	if attempt != nil {
		<-attempt.done
	}

	p.closeSessions(ctx, sessions, false)
	p.recordCleanup(closeConnection(conn))
	if interceptor != nil {
		p.recordCleanup(interceptor.Close())
	}

	if watchDone != nil {
		<-watchDone
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	return p.cleanupErr
}
