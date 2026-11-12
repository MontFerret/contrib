package cdp

import (
	"context"
	stderrors "errors"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/rpcc"
	"github.com/pkg/errors"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/ferret/v2/pkg/logging"
)

const DriverName = "cdp"
const BlankPageURL = "about:blank"

var defaultViewport = &drivers.Viewport{
	Width:  1600,
	Height: 900,
}

type Driver struct {
	dev     *devtool.DevTools
	options *Options
	policy  *browserPolicy
}

func New(opts ...Option) *Driver {
	drv := new(Driver)
	drv.options = NewOptions(opts)
	drv.dev = devtool.New(drv.options.Address)

	if drv.options.AllowedOrigins != nil {
		drv.policy = newBrowserPolicy(drv.options.AllowedOrigins)
	}

	return drv
}

func (drv *Driver) Name() string {
	return drv.options.Name
}

func (drv *Driver) Open(ctx context.Context, params drivers.Params) (drivers.HTMLPage, error) {
	logger := logging.From(ctx)

	owned, ctx, err := drv.beginPolicySession(ctx)
	if err != nil {
		return nil, err
	}

	sessions, err := drv.createSessionManager(ctx, params.KeepCookies, owned)
	if err != nil {
		logger.Error().
			Err(err).
			Str("driver", drv.options.Name).
			Msg("failed to create a new connection")

		return nil, err
	}

	page, err := loadHTMLPage(ctx, sessions, drv.setDefaultParams(params), owned)

	return owned.finish(page, err)
}

func (drv *Driver) Parse(ctx context.Context, params drivers.ParseParams) (drivers.HTMLPage, error) {
	logger := logging.From(ctx)

	owned, ctx, err := drv.beginPolicySession(ctx)
	if err != nil {
		return nil, err
	}

	sessions, err := drv.createSessionManager(ctx, true, owned)
	if err != nil {
		logger.Error().
			Err(err).
			Str("driver", drv.options.Name).
			Msg("failed to create a new connection")

		return nil, err
	}

	page, err := loadHTMLPageWithContent(ctx, sessions, drv.setDefaultParams(drivers.Params{
		URL:         BlankPageURL,
		UserAgent:   "",
		KeepCookies: params.KeepCookies,
		Cookies:     params.Cookies,
		Headers:     params.Headers,
		Viewport:    params.Viewport,
	}), params.Content, owned)

	return owned.finish(page, err)
}

// Close releases the browser policy connection and invalidates owned pages.
// Restricted drivers return unresolved cleanup failures after bounded attempts;
// the endpoint may still contain live targets when cleanup fails.
// Without an allowlist, page lifecycles remain owned by their callers.
func (drv *Driver) Close() error {
	if drv.policy != nil {
		return drv.policy.Close()
	}

	return nil
}

func (drv *Driver) beginPolicySession(ctx context.Context) (*policySession, context.Context, error) {
	if drv.policy == nil {
		owned := newPageSession(ctx, nil, drv)

		return owned, owned.ctx, nil
	}

	owned, err := drv.policy.begin(ctx, drv)
	if err != nil {
		return nil, ctx, err
	}

	return owned, owned.ctx, nil
}

func (drv *Driver) createSessionManager(ctx context.Context, keepCookies bool, owned *policySession) (manager *cdpsession.Manager, err error) {
	defer func() {
		if err != nil {
			err = stderrors.Join(err, owned.Close())
		}
	}()

	ver, err := drv.dev.Version(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get browser endpoint")
	}

	conn, err := cdpsession.DialFlat(ctx, ver.WebSocketDebuggerURL, drv.dialOptions()...)
	if err != nil {
		return nil, errors.Wrap(err, "initialize driver")
	}

	browserClient := cdp.NewClient(conn.Conn)
	if err := owned.setConnection(conn.Conn, browserClient); err != nil {
		return nil, err
	}

	target, err := owned.createTarget(ctx, browserClient, !drv.options.KeepCookies && !keepCookies)
	if err != nil {
		return nil, errors.Wrap(err, "create a browser target")
	}

	sessions, err := cdpsession.NewFlat(ctx, conn, target.TargetID)
	if err != nil {
		return nil, errors.Wrap(err, "establish a new connection")
	}

	if err := owned.setSessions(sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (drv *Driver) setDefaultParams(params drivers.Params) drivers.Params {
	if params.Viewport == nil {
		params.Viewport = defaultViewport
	}

	return drivers.SetDefaultParams(drv.options.Options, params)
}

func (drv *Driver) openBrowser(ctx context.Context) (*rpcc.Conn, *cdp.Client, error) {
	ver, err := drv.dev.Version(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize driver")
	}

	conn, err := rpcc.DialContext(ctx, ver.WebSocketDebuggerURL, drv.dialOptions()...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize driver")
	}

	return conn, cdp.NewClient(conn), nil
}

func (drv *Driver) dialOptions() []rpcc.DialOption {
	dialOpts := make([]rpcc.DialOption, 0, 2)
	if drv.options.Connection != nil {
		if drv.options.Connection.BufferSize > 0 {
			dialOpts = append(dialOpts, rpcc.WithWriteBufferSize(drv.options.Connection.BufferSize))
		}

		if drv.options.Connection.Compression {
			dialOpts = append(dialOpts, rpcc.WithCompression())
		}
	}

	return dialOpts
}
