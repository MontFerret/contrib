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
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/internal/cleanup"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
)

// policySession owns a driver-created page and its pending initialization.
// Restricted drivers also register this owner for browser-wide invalidation.
type policySession struct {
	mu          sync.Mutex
	policy      *browserPolicy
	driver      *Driver
	ctx         context.Context
	cancel      context.CancelFunc
	lifetime    context.Context
	invalidate  context.CancelFunc
	conn        *rpcc.Conn
	client      *cdp.Client
	sessions    *cdpsession.Manager
	page        *HTMLPage
	targetID    target.ID
	contextID   browser.ContextID
	closed      bool
	initialized chan struct{}
	creationErr error
	cleanup     cleanup.State
}

func newPageSession(ctx context.Context, policy *browserPolicy, driver *Driver) *policySession {
	operation, cancel := context.WithCancel(ctx)
	lifetime, invalidate := context.WithCancel(context.Background())

	return &policySession{policy: policy, driver: driver, ctx: operation, cancel: cancel, lifetime: lifetime, invalidate: invalidate, initialized: make(chan struct{})}
}

func (s *policySession) setConnection(conn *rpcc.Conn, client *cdp.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.conn, s.client = conn, client
	if s.closed {
		return errors.New("CDP session invalidated by browser policy shutdown")
	}

	return nil
}

func (s *policySession) setSessions(manager *cdpsession.Manager) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = manager
	if s.closed {
		return errors.New("CDP session invalidated by browser policy shutdown")
	}

	return nil
}

func (s *policySession) createTarget(ctx context.Context, client *cdp.Client, isolated bool) (*target.CreateTargetReply, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	args := target.NewCreateTargetArgs(BlankPageURL)
	if isolated {
		reply, err := client.Target.CreateBrowserContext(ctx, target.NewCreateBrowserContextArgs().SetDisposeOnDetach(true))
		if err != nil {
			s.recordCreationError(err)

			return nil, err
		}

		s.contextID = reply.BrowserContextID
		args.SetBrowserContextID(s.contextID)
	}

	reply, err := client.Target.CreateTarget(ctx, args)
	if err != nil {
		if s.contextID == "" {
			s.recordCreationError(err)
		}

		return nil, err
	}

	s.targetID = reply.TargetID

	return reply, nil
}

// A failed transport can lose a creation reply after Chromium acted on it.
func (s *policySession) recordCreationError(err error) {
	var rejected *rpcc.ResponseError
	if !errors.As(err, &rejected) {
		s.creationErr = fmt.Errorf("cannot confirm CDP resource creation or disposal: %w", err)
	}
}

func (s *policySession) finish(page *HTMLPage, err error) (drivers.HTMLPage, error) {
	s.mu.Lock()
	if err == nil {
		err = s.ctx.Err()
		if err == nil && s.closed {
			err = errors.New("CDP session invalidated during page initialization")
		}
	}

	s.mu.Unlock()
	s.cancel()
	close(s.initialized)

	if err == nil {
		return page, nil
	}

	if s.managed() {
		s.invalidate()

		return nil, err
	}

	return nil, errors.Join(err, s.Close())
}

func (s *policySession) setPage(page *HTMLPage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.page = page
	page.policySession = s
	if s.closed {
		return errors.New("CDP session invalidated during page initialization")
	}

	return nil
}

func (s *policySession) managed() bool {
	return s.driver != nil && s.driver.options.managedPageCleanup
}

func (s *policySession) Close() error {
	ctx, cancel := defaultCleanupContext()
	defer cancel()

	return s.CloseContext(ctx)
}

func (s *policySession) CloseContext(ctx context.Context) error {
	s.invalidate()
	s.cancel()
	s.mu.Lock()
	s.closed = true
	sessions := s.sessions
	s.mu.Unlock()

	if sessions != nil {
		sessions.StopRouting()
	}

	return s.cleanup.Run(ctx, s.close)
}

func (s *policySession) close(ctx context.Context) error {
	if s.initialized != nil {
		<-s.initialized
	}

	err := errors.Join(s.creationErr, s.closeTarget(ctx), closeConnection(s.conn))
	if s.sessions != nil {
		// Closing the browser socket releases commands holding page locks.
		managerErr := s.sessions.CloseContext(ctx)
		if ctx.Err() != nil {
			managerErr = s.sessions.CloseContext(context.Background())
		}

		err = errors.Join(err, managerErr)
	}

	if s.page != nil {
		err = errors.Join(err, s.page.closeResources())
	}

	if s.policy != nil {
		s.policy.release(s, err)
	}

	return err
}

func (s *policySession) closeTarget(ctx context.Context) error {
	if s.targetID == "" && s.contextID == "" {
		return nil
	}

	var primary *cdp.Client
	if s.conn != nil && s.conn.Context().Err() == nil {
		primary = s.client
	}

	var result error
	if primary != nil && ctx.Err() == nil {
		primaryCtx, stop := primaryCleanupContext(ctx)
		result = s.disposeTarget(primaryCtx, primary)
		stop()

		if result == nil {
			return nil
		}
	}

	if result == nil {
		result = errors.New("browser cleanup connection unavailable")
	}

	var fallback *cdp.Client
	if s.driver != nil {
		// Release the context owner before confirming disposal on another
		// socket, so Chromium can apply disposeOnDetach.
		result = errors.Join(result, closeConnection(s.conn))
		if ctx.Err() == nil {
			conn, client, err := s.driver.openBrowser(ctx)
			if err == nil {
				fallback = client
				defer closeConnection(conn)
			} else {
				result = errors.Join(result, err)
			}
		}
	}

	if fallback != nil && ctx.Err() == nil {
		err := s.disposeTarget(ctx, fallback)
		if err == nil {
			return nil
		}

		result = errors.Join(result, err)
	}

	if fallback == nil && s.driver == nil {
		fallback = primary
	}

	if fallback != nil && ctx.Err() == nil && (s.contextID != "" || s.sessions == nil) {
		gone, err := s.targetGone(ctx, fallback)
		if err == nil && gone {
			return nil
		}

		result = errors.Join(result, err)
	}

	result = errors.Join(result, ctx.Err())
	if s.contextID != "" {
		return fmt.Errorf("dispose CDP browser context %s: %w", s.contextID, result)
	}

	return fmt.Errorf("close CDP target %s: %w", s.targetID, result)
}

func (s *policySession) disposeTarget(ctx context.Context, client *cdp.Client) error {
	if client == nil {
		return errors.New("browser cleanup connection unavailable")
	}

	if s.contextID != "" {
		return client.Target.DisposeBrowserContext(ctx, target.NewDisposeBrowserContextArgs(s.contextID))
	}

	if s.sessions != nil {
		return s.sessions.ClosePageTargets(ctx, client)
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
