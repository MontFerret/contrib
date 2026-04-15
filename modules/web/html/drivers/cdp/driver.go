package cdp

import (
	"context"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/target"
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
}

func New(opts ...Option) *Driver {
	drv := new(Driver)
	drv.options = NewOptions(opts)
	drv.dev = devtool.New(drv.options.Address)

	return drv
}

func (drv *Driver) Name() string {
	return drv.options.Name
}

func (drv *Driver) Open(ctx context.Context, params drivers.Params) (drivers.HTMLPage, error) {
	logger := logging.From(ctx)

	sessions, err := drv.createSessionManager(ctx, params.KeepCookies)

	if err != nil {
		logger.Error().
			Err(err).
			Str("driver", drv.options.Name).
			Msg("failed to create a new connection")

		return nil, err
	}

	return LoadHTMLPage(ctx, sessions, drv.setDefaultParams(params))
}

func (drv *Driver) Parse(ctx context.Context, params drivers.ParseParams) (drivers.HTMLPage, error) {
	logger := logging.From(ctx)

	sessions, err := drv.createSessionManager(ctx, true)

	if err != nil {
		logger.Error().
			Err(err).
			Str("driver", drv.options.Name).
			Msg("failed to create a new connection")

		return nil, err
	}

	return LoadHTMLPageWithContent(ctx, sessions, drv.setDefaultParams(drivers.Params{
		URL:         BlankPageURL,
		UserAgent:   "",
		KeepCookies: params.KeepCookies,
		Cookies:     params.Cookies,
		Headers:     params.Headers,
		Viewport:    params.Viewport,
	}), params.Content)
}

func (drv *Driver) Close() error {
	return nil
}

func (drv *Driver) createSessionManager(ctx context.Context, keepCookies bool) (*cdpsession.Manager, error) {
	browserConn, browserClient, err := drv.openBrowser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "initialize driver")
	}

	createTargetArgs := target.NewCreateTargetArgs(BlankPageURL)
	if !drv.options.KeepCookies && !keepCookies {
		ctxReply, err := browserClient.Target.CreateBrowserContext(
			ctx,
			target.NewCreateBrowserContextArgs().SetDisposeOnDetach(true),
		)
		if err != nil {
			_ = browserConn.Close()

			return nil, err
		}

		createTargetArgs.SetBrowserContextID(ctxReply.BrowserContextID)
	}

	createTarget, err := browserClient.Target.CreateTarget(ctx, createTargetArgs)
	if err != nil {
		_ = browserConn.Close()

		return nil, errors.Wrap(err, "create a browser target")
	}

	sessions, err := cdpsession.New(ctx, browserConn, browserClient, createTarget.TargetID)
	if err != nil {
		_ = browserConn.Close()

		return nil, errors.Wrap(err, "establish a new connection")
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

	dialOpts := make([]rpcc.DialOption, 0, 2)
	if drv.options.Connection != nil {
		if drv.options.Connection.BufferSize > 0 {
			dialOpts = append(dialOpts, rpcc.WithWriteBufferSize(drv.options.Connection.BufferSize))
		}

		if drv.options.Connection.Compression {
			dialOpts = append(dialOpts, rpcc.WithCompression())
		}
	}

	conn, err := rpcc.DialContext(ctx, ver.WebSocketDebuggerURL, dialOpts...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to initialize driver")
	}

	return conn, cdp.NewClient(conn), nil
}
