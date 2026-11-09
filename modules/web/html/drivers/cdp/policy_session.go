package cdp

import (
	"context"
	"errors"
	"fmt"
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
		if cleanupErr := s.Close(); cleanupErr != nil && !errors.Is(err, cleanupErr) {
			err = errors.Join(err, cleanupErr)
		}

		return nil, err
	}

	s.mu.Lock()
	if s.closed || s.ctx.Err() != nil {
		failure := s.ctx.Err()
		if failure == nil {
			failure = errors.New("CDP session invalidated during page initialization")
		}

		s.mu.Unlock()

		return nil, errors.Join(failure, s.Close())
	}

	s.mu.Unlock()

	return page, nil
}

func (s *policySession) setPage(page *HTMLPage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errors.New("CDP session invalidated during page initialization")
	}

	s.page = page
	page.policySession = s

	return nil
}

func (s *policySession) Close() error {
	s.closeOnce.Do(func() {
		s.invalidate()
		s.cancel()
		s.mu.Lock()
		s.closed = true
		conn, page, sessions := s.conn, s.page, s.sessions
		s.mu.Unlock()

		s.closeErr = s.closeTarget()
		if conn != nil && conn.Context().Err() == nil {
			s.closeErr = errors.Join(s.closeErr, conn.Close())
		}

		if sessions != nil {
			// Cancel tunneled clients before taking page locks held by commands.
			s.closeErr = errors.Join(s.closeErr, sessions.Close())
		}

		if page != nil {
			s.closeErr = errors.Join(s.closeErr, page.closeResources())
		}

		s.policy.release(s, s.closeErr)
	})

	return s.closeErr
}

func (s *policySession) closeTarget() error {
	if s.targetID == "" && s.contextID == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), policyCleanupTimeout)
	defer cancel()

	var primary *cdp.Client
	if s.conn != nil && s.conn.Context().Err() == nil {
		primary = s.client
	}

	primaryCtx, stop := context.WithTimeout(ctx, policyCleanupTimeout/2)
	err := s.disposeTarget(primaryCtx, primary)
	stop()
	if err == nil {
		return nil
	}

	fallback := s.policy.cleanupClient()
	if fallback != nil {
		fallbackErr := s.disposeTarget(ctx, fallback)
		if fallbackErr == nil {
			return nil
		}

		err = errors.Join(err, fallbackErr)
	} else {
		fallback = primary
	}

	// A command can fail because a concurrent close already removed the target.
	// Confirm absence with protocol state, rather than matching error strings.
	if fallback != nil {
		gone, confirmationErr := s.targetGone(ctx, fallback)
		if confirmationErr == nil && gone {
			return nil
		}

		err = errors.Join(err, confirmationErr)
	}

	if s.contextID != "" {
		return fmt.Errorf("dispose CDP browser context %s: %w", s.contextID, err)
	}

	return fmt.Errorf("close CDP target %s: %w", s.targetID, err)
}

func (s *policySession) disposeTarget(ctx context.Context, client *cdp.Client) error {
	if client == nil {
		return errors.New("browser cleanup connection unavailable")
	}

	if s.contextID != "" {
		return client.Target.DisposeBrowserContext(ctx, target.NewDisposeBrowserContextArgs(s.contextID))
	}

	_, err := client.Target.CloseTarget(ctx, target.NewCloseTargetArgs(s.targetID))

	return err
}

func (s *policySession) targetGone(ctx context.Context, client *cdp.Client) (bool, error) {
	if s.contextID != "" {
		reply, err := client.Target.GetBrowserContexts(ctx)
		if err != nil {
			return false, err
		}

		for _, id := range reply.BrowserContextIDs {
			if id == s.contextID {
				return false, nil
			}
		}

		return true, nil
	}

	reply, err := client.Target.GetTargets(ctx, target.NewGetTargetsArgs())
	if err != nil {
		return false, err
	}

	for _, info := range reply.TargetInfos {
		if info.TargetID == s.targetID {
			return false, nil
		}
	}

	return true, nil
}
