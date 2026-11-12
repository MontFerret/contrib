package cdp

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	cdpnet "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
	"github.com/MontFerret/ferret/v2/pkg/logging"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	HTMLPageEvent string

	HTMLPage struct {
		logger        zerolog.Logger
		client        *cdp.Client
		sessions      *cdpsession.Manager
		network       *cdpnet.Manager
		dom           *dom.Manager
		initScript    *drivers.InitScript
		policySession *policySession
		mu            sync.Mutex
		closed        runtime.Boolean
		closeOnce     sync.Once
		closeErr      error
	}
)

const (
	frameRefreshInterval    = 25 * time.Millisecond
	maxFrameRefreshAttempts = 20
)

func LoadHTMLPage(
	ctx context.Context,
	sessions *cdpsession.Manager,
	params drivers.Params,
) (*HTMLPage, error) {
	return loadHTMLPage(ctx, sessions, params, nil)
}

func loadHTMLPage(ctx context.Context, sessions *cdpsession.Manager, params drivers.Params, owned *policySession) (p *HTMLPage, err error) {
	logger := logging.From(ctx)
	initScript, err := drivers.NormalizeInitScript(params.InitScript)
	if err != nil {
		return nil, err
	}

	if sessions == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "sessions")
	}

	root := sessions.Root()
	if root == nil || root.CDP == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "root session")
	}

	client := root.CDP
	var netManager *cdpnet.Manager
	var domManager *dom.Manager

	defer func() {
		if err != nil {
			if p != nil {
				err = errors.Join(err, p.Close())

				return
			}

			if domManager != nil {
				if closeErr := domManager.Close(); closeErr != nil {
					logger.Error().Err(closeErr).Msg("failed to close DOM manager")

					if owned != nil || sessions.IsFlat() {
						err = errors.Join(err, closeErr)
					}
				}
			}

			if netManager != nil {
				if closeErr := netManager.Close(); closeErr != nil {
					logger.Error().Err(closeErr).Msg("failed to close network manager")

					if owned != nil || sessions.IsFlat() {
						err = errors.Join(err, closeErr)
					}
				}
			}

			if owned != nil {
				err = errors.Join(err, owned.Close())

				return
			}

			cleanupCtx, cancel := context.WithTimeout(context.Background(), policyCleanupTimeout)
			defer cancel()
			if closeErr := sessions.ClosePageTargets(cleanupCtx, nil); closeErr != nil {
				logger.Error().Err(closeErr).Msg("failed to close page")
				if sessions.IsFlat() {
					err = errors.Join(err, closeErr)
				}
			}

			if closeErr := sessions.Close(); closeErr != nil {
				logger.Error().Err(closeErr).Msg("failed to close session manager")
				if sessions.IsFlat() {
					err = errors.Join(err, closeErr)
				}
			}
		}
	}()

	if err = enableFeatures(ctx, client, params); err != nil {
		return nil, err
	}

	netOpts := cdpnet.Options{
		Headers: params.Headers,
	}

	if params.Cookies != nil && len(params.Cookies.Data) > 0 {
		netOpts.Cookies = make(map[string]*drivers.HTTPCookies)
		netOpts.Cookies[params.URL] = params.Cookies
	}

	if params.Ignore != nil && len(params.Ignore.Resources) > 0 {
		netOpts.Filter = &cdpnet.Filter{
			Patterns: params.Ignore.Resources,
		}
	}

	netManager, err = cdpnet.New(
		logger,
		client,
		sessions,
		netOpts,
	)

	if err != nil {
		return nil, err
	}

	mouse := input.NewMouse(client)
	keyboard := input.NewKeyboard(client)

	domManager, err = dom.New(
		logger,
		client,
		mouse,
		keyboard,
	)

	if err != nil {
		return nil, err
	}

	domManager.SetFrameSessions(sessions)

	p = NewHTMLPage(
		logger,
		client,
		sessions,
		netManager,
		domManager,
	)
	p.initScript = initScript

	if owned != nil {
		if err = owned.setPage(p); err != nil {
			return p, err
		}
	}

	if err = p.registerInitScript(ctx); err != nil {
		return p, err
	}

	if params.URL != BlankPageURL && params.URL != "" {
		err = p.Navigate(ctx, runtime.NewString(params.URL))
	} else {
		err = p.loadMainFrame(ctx)
	}

	if err != nil {
		return p, err
	}

	return p, nil
}

func LoadHTMLPageWithContent(
	ctx context.Context,
	sessions *cdpsession.Manager,
	params drivers.Params,
	content []byte,
) (*HTMLPage, error) {
	return loadHTMLPageWithContent(ctx, sessions, params, content, nil)
}

func loadHTMLPageWithContent(ctx context.Context, sessions *cdpsession.Manager, params drivers.Params, content []byte, owned *policySession) (p *HTMLPage, err error) {
	logger := logging.From(ctx)
	p, err = loadHTMLPage(ctx, sessions, params, owned)

	if err != nil {
		return nil, err
	}

	defer func(page *HTMLPage) {
		if err != nil {
			if e := page.Close(); e != nil {
				logger.Error().Err(e).Msg("failed to close page")
				err = errors.Join(err, e)
			}
		}
	}(p)

	frameID := p.getCurrentDocument().Frame().Frame.ID
	err = p.client.Page.SetDocumentContent(ctx, page.NewSetDocumentContentArgs(frameID, string(content)))

	if err != nil {
		return nil, runtime.Error(err, "set document content")
	}

	// Remove prev frames (from a blank page)
	prev := p.dom.GetMainFrame()
	err = p.dom.RemoveFrameRecursively(prev.Frame().Frame.ID)

	if err != nil {
		return nil, err
	}

	err = p.loadMainFrame(ctx)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func NewHTMLPage(
	logger zerolog.Logger,
	client *cdp.Client,
	sessions *cdpsession.Manager,
	netManager *cdpnet.Manager,
	domManager *dom.Manager,
) *HTMLPage {
	p := new(HTMLPage)
	p.closed = runtime.False
	p.logger = logutil.WithComponent(logger.With(), "cdp_page").Logger()
	p.client = client
	p.sessions = sessions
	p.network = netManager
	p.dom = domManager

	return p
}
