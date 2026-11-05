package cdp

import (
	"context"
	"errors"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/browser"
	"github.com/mafredri/cdp/protocol/target"
	"github.com/mafredri/cdp/rpcc"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

// policySession includes pending page initialization so policy loss can cancel
// it before a page is published to the caller.
type policySession struct {
	mu         sync.Mutex
	policy     *browserPolicy
	ctx        context.Context
	cancel     context.CancelFunc
	lifetime   context.Context
	invalidate context.CancelFunc
	conn       *rpcc.Conn
	client     *cdp.Client
	sessions   *cdpsession.Manager
	page       *HTMLPage
	targetID   target.ID
	contextID  browser.ContextID
	closed     bool
	closeOnce  sync.Once
	closeErr   error
}

func (s *policySession) setConnection(conn *rpcc.Conn, client *cdp.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		_ = conn.Close()

		return errors.New("CDP session invalidated by browser policy shutdown")
	}

	s.conn, s.client = conn, client

	return nil
}

func (s *policySession) setSessions(manager *cdpsession.Manager) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		_ = manager.Close()

		return errors.New("CDP session invalidated by browser policy shutdown")
	}

	s.sessions = manager

	return nil
}

func (s *policySession) createTarget(ctx context.Context, client *cdp.Client, isolated bool) (*target.CreateTargetReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("CDP session invalidated before target creation")
	}

	args := target.NewCreateTargetArgs(BlankPageURL)
	if isolated {
		reply, err := client.Target.CreateBrowserContext(ctx, target.NewCreateBrowserContextArgs().SetDisposeOnDetach(true))
		if err != nil {
			return nil, err
		}

		s.contextID = reply.BrowserContextID
		args.SetBrowserContextID(s.contextID)
	}

	reply, err := client.Target.CreateTarget(ctx, args)
	if err != nil {
		return nil, err
	}

	s.targetID = reply.TargetID

	return reply, nil
}

func (s *policySession) finish(page *HTMLPage, err error) (drivers.HTMLPage, error) {
	if s == nil {
		return page, err
	}

	defer s.cancel()

	if err != nil {
		_ = s.Close()

		return nil, err
	}

	s.mu.Lock()
	if s.closed || s.ctx.Err() != nil {
		failure := s.ctx.Err()
		if failure == nil {
			failure = errors.New("CDP session invalidated during page initialization")
		}

		s.mu.Unlock()
		_ = page.Close()
		_ = s.Close()

		return nil, failure
	}

	s.page = page
	page.policySession = s
	s.mu.Unlock()

	return page, nil
}

func (s *policySession) release() {
	s.policy.release(s)
}

func (s *policySession) Close() error {
	s.closeOnce.Do(func() {
		s.invalidate()
		s.cancel()
		s.mu.Lock()
		s.closed = true
		conn, client, page, sessions := s.conn, s.client, s.page, s.sessions
		targetID, contextID := s.targetID, s.contextID
		s.mu.Unlock()

		// Close browser targets through the independent browser channel first.
		// Page methods may hold their mutex while waiting for a CDP response.
		ctx, cancel := context.WithTimeout(context.Background(), policyCleanupTimeout)
		defer cancel()

		if client != nil && conn.Context().Err() == nil {
			if contextID != "" {
				_ = client.Target.DisposeBrowserContext(ctx, target.NewDisposeBrowserContextArgs(contextID))
			} else if targetID != "" {
				_, _ = client.Target.CloseTarget(ctx, target.NewCloseTargetArgs(targetID))
			}
		}

		if conn != nil {
			s.closeErr = conn.Close()
		}

		if sessions != nil {
			// Closing the browser connection alone does not cancel the tunneled
			// clients. Unblock their pending commands before acquiring page locks.
			_ = sessions.Close()
		}

		if page != nil {
			s.closeErr = errors.Join(s.closeErr, page.Close())
		}

		s.release()
	})

	return s.closeErr
}
