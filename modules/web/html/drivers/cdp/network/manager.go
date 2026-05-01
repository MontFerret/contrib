package network

import (
	"context"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/contrib/modules/web/html/internal/logutil"
)

const BlankPageURL = "about:blank"

type (
	FrameLoadedListener = func(ctx context.Context, frame page.Frame)

	Manager struct {
		logger           zerolog.Logger
		client           *cdp.Client
		sessions         *cdpsession.Manager
		headers          *drivers.HTTPHeaders
		interceptor      *Interceptor
		stop             context.CancelFunc
		response         *sync.Map
		responseWatchers map[string]network.ResponseReceivedClient
		responseListener cdpsession.ListenerID
		mu               sync.RWMutex
		responseMu       sync.Mutex
	}
)

func New(
	logger zerolog.Logger,
	client *cdp.Client,
	sessions *cdpsession.Manager,
	options Options,
) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := new(Manager)
	m.logger = logutil.WithComponent(logger.With(), "network_manager").Logger()
	m.client = client
	m.sessions = sessions
	m.headers = drivers.NewHTTPHeaders()
	m.stop = cancel
	m.response = new(sync.Map)
	m.responseWatchers = make(map[string]network.ResponseReceivedClient)

	var err error

	defer func() {
		if err != nil {
			m.stop()
		}
	}()

	if options.Filter != nil && len(options.Filter.Patterns) > 0 {
		m.interceptor = NewInterceptor(logger, client)

		if err := m.interceptor.AddFilter("resources", options.Filter); err != nil {
			return nil, err
		}

		if err = m.interceptor.Run(ctx); err != nil {
			return nil, err
		}
	}

	if len(options.Cookies) > 0 {
		for url, cookies := range options.Cookies {
			err = m.setCookiesInternal(ctx, url, cookies)

			if err != nil {
				return nil, err
			}
		}
	}

	if options.Headers != nil && len(options.Headers.Data) > 0 {
		err = m.setHeadersInternal(ctx, options.Headers)

		if err != nil {
			return nil, err
		}
	}

	if err = m.startResponseWatcher(ctx); err != nil {
		return nil, err
	}

	return m, nil
}
