package cdp

import (
	"context"
	"errors"
	"hash/fnv"
	"regexp"
	"sync"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/rs/zerolog"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/dom"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/eval"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/events"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/input"
	cdpnet "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/network"
	cdpsession "github.com/MontFerret/contrib/modules/web/html/drivers/cdp/session"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/templates"
	"github.com/MontFerret/contrib/modules/web/html/drivers/cdp/utils"
	"github.com/MontFerret/contrib/modules/web/html/drivers/common"
	"github.com/MontFerret/ferret/v2/pkg/logging"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

type (
	HTMLPageEvent string

	HTMLPage struct {
		logger   zerolog.Logger
		client   *cdp.Client
		sessions *cdpsession.Manager
		network  *cdpnet.Manager
		dom      *dom.Manager
		mu       sync.Mutex
		closed   runtime.Boolean
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
) (p *HTMLPage, err error) {
	logger := logging.From(ctx)

	if sessions == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "sessions")
	}

	root := sessions.Root()
	if root == nil || root.CDP == nil {
		return nil, runtime.Error(runtime.ErrMissedArgument, "root session")
	}

	client := root.CDP
	if err := enableFeatures(ctx, client, params); err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err := client.Page.Close(context.Background()); err != nil {
				logger.Error().Err(err).Msg("failed to close page")
			}

			if err := sessions.Close(); err != nil {
				logger.Error().Err(err).Msg("failed to close session manager")
			}
		}
	}()

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

	netManager, err := cdpnet.New(
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

	domManager, err := dom.New(
		logger,
		client,
		mouse,
		keyboard,
	)

	if err != nil {
		return nil, err
	}

	p = NewHTMLPage(
		logger,
		client,
		sessions,
		netManager,
		domManager,
	)

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
) (p *HTMLPage, err error) {
	logger := logging.From(ctx)
	p, err = LoadHTMLPage(ctx, sessions, params)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if e := p.Close(); e != nil {
				logger.Error().Err(e).Msg("failed to close page")
			}
		}
	}()

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
	p.logger = common.LoggerWithName(logger.With(), "cdp_page").Logger()
	p.client = client
	p.sessions = sessions
	p.network = netManager
	p.dom = domManager

	return p
}

func (p *HTMLPage) MarshalJSON() ([]byte, error) {
	return p.getCurrentDocument().MarshalJSON()
}

func (p *HTMLPage) Type() runtime.Type {
	return drivers.HTMLPageType
}

func (p *HTMLPage) String() string {
	return p.getCurrentDocument().GetURL().String()
}

func (p *HTMLPage) Compare(other runtime.Value) int {
	cdpPage, ok := other.(*HTMLPage)

	if !ok {
		return drivers.CompareTypes(p, other)
	}

	return p.getCurrentDocument().GetURL().Compare(cdpPage.GetURL())
}

func (p *HTMLPage) Unwrap() any {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p
}

func (p *HTMLPage) Hash() uint64 {
	h := fnv.New64a()

	h.Write([]byte("CDP"))
	h.Write([]byte(p.Type().String()))
	h.Write([]byte(":"))
	h.Write([]byte(p.getCurrentDocument().GetURL()))

	return h.Sum64()
}

func (p *HTMLPage) Copy() runtime.Value {
	return runtime.None
}

func (p *HTMLPage) Get(ctx context.Context, key runtime.Value) (runtime.Value, error) {
	return common.GetInPage(ctx, key, p)
}

func (p *HTMLPage) Set(ctx context.Context, key runtime.Value, value runtime.Value) error {
	return common.SetInPage(ctx, key, p, value)
}

func (p *HTMLPage) Iterate(ctx context.Context) (runtime.Iterator, error) {
	return p.getCurrentDocument().Iterate(ctx)
}

func (p *HTMLPage) Length(ctx context.Context) (runtime.Int, error) {
	return p.getCurrentDocument().Length(ctx)
}

func (p *HTMLPage) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var url string
	frame := p.dom.GetMainFrame()

	if frame != nil {
		url = frame.GetURL().String()
	}

	p.closed = runtime.True

	err := p.dom.Close()

	if err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close dom manager")
	}

	err = p.network.Close()

	if err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close network manager")
	}

	err = p.client.Page.Close(context.Background())

	if err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close browser page")
	}

	if err := p.sessions.Close(); err != nil {
		p.logger.Warn().
			Str("url", url).
			Err(err).
			Msg("failed to close session manager")
	}

	return nil
}

func (p *HTMLPage) IsClosed() runtime.Boolean {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.closed
}

func (p *HTMLPage) GetURL() runtime.String {
	res, err := p.getCurrentDocument().Eval().EvalValue(context.Background(), templates.GetURL())

	if err == nil {
		return runtime.ToString(res)
	}

	p.logger.Warn().
		Err(err).
		Msg("failed to retrieve URL")

	return p.getCurrentDocument().GetURL()
}

func (p *HTMLPage) GetMainFrame() drivers.HTMLDocument {
	return p.getCurrentDocument()
}

func (p *HTMLPage) GetFrames(ctx context.Context) (runtime.List, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.loadFrames(ctx)
}

func (p *HTMLPage) GetFrame(ctx context.Context, idx runtime.Int) (runtime.Value, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	frames, err := p.loadFrames(ctx)

	if err != nil {
		return runtime.None, err
	}

	return frames.At(ctx, idx)
}

func (p *HTMLPage) GetCookies(ctx context.Context) (*drivers.HTTPCookies, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.network.GetCookies(ctx, p.getCurrentDocument().GetURL().String())
}

func (p *HTMLPage) SetCookies(ctx context.Context, cookies *drivers.HTTPCookies) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.network.SetCookies(ctx, p.getCurrentDocument().GetURL().String(), cookies)
}

func (p *HTMLPage) DeleteCookies(ctx context.Context, cookies *drivers.HTTPCookies) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.network.DeleteCookies(ctx, p.getCurrentDocument().GetURL().String(), cookies)
}

func (p *HTMLPage) GetResponse(ctx context.Context) (drivers.HTTPResponse, error) {
	doc := p.getCurrentDocument()

	if doc == nil {
		return drivers.HTTPResponse{}, nil
	}

	return p.network.GetResponse(ctx, doc.Frame().Frame.ID)
}

func (p *HTMLPage) PrintToPDF(ctx context.Context, params drivers.PDFParams) (runtime.Binary, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	args := page.NewPrintToPDFArgs()
	args.
		SetLandscape(bool(params.Landscape)).
		SetDisplayHeaderFooter(bool(params.DisplayHeaderFooter)).
		SetPrintBackground(bool(params.PrintBackground)).
		SetPreferCSSPageSize(bool(params.PreferCSSPageSize))

	if params.Scale > 0 {
		args.SetScale(float64(params.Scale))
	}

	if params.PaperWidth > 0 {
		args.SetPaperWidth(float64(params.PaperWidth))
	}

	if params.PaperHeight > 0 {
		args.SetPaperHeight(float64(params.PaperHeight))
	}

	if params.MarginTop > 0 {
		args.SetMarginTop(float64(params.MarginTop))
	}

	if params.MarginBottom > 0 {
		args.SetMarginBottom(float64(params.MarginBottom))
	}

	if params.MarginRight > 0 {
		args.SetMarginRight(float64(params.MarginRight))
	}

	if params.MarginLeft > 0 {
		args.SetMarginLeft(float64(params.MarginLeft))
	}

	if params.PageRanges != runtime.EmptyString {
		args.SetPageRanges(string(params.PageRanges))
	}

	if params.HeaderTemplate != runtime.EmptyString {
		args.SetHeaderTemplate(string(params.HeaderTemplate))
	}

	if params.FooterTemplate != runtime.EmptyString {
		args.SetFooterTemplate(string(params.FooterTemplate))
	}

	reply, err := p.client.Page.PrintToPDF(ctx, args)

	if err != nil {
		return runtime.NewBinary([]byte{}), err
	}

	return runtime.NewBinary(reply.Data), nil
}

func (p *HTMLPage) CaptureScreenshot(ctx context.Context, params drivers.ScreenshotParams) (runtime.Binary, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	metrics, err := p.client.Page.GetLayoutMetrics(ctx)

	if err != nil {
		return runtime.NewBinary(nil), err
	}

	if params.Format == drivers.ScreenshotFormatJPEG && (params.Quality < 0 || params.Quality > 100) {
		params.Quality = 100
	}

	if params.X < 0 {
		params.X = 0
	}

	if params.Y < 0 {
		params.Y = 0
	}

	clientWidth, clientHeight := utils.GetLayoutViewportWH(metrics)

	if params.Width <= 0 {
		params.Width = runtime.Float(clientWidth) - params.X
	}

	if params.Height <= 0 {
		params.Height = runtime.Float(clientHeight) - params.Y
	}

	clip := page.Viewport{
		X:      float64(params.X),
		Y:      float64(params.Y),
		Width:  float64(params.Width),
		Height: float64(params.Height),
		Scale:  1.0,
	}

	format := string(params.Format)
	quality := int(params.Quality)
	args := page.CaptureScreenshotArgs{
		Format:  &format,
		Quality: &quality,
		Clip:    &clip,
	}

	reply, err := p.client.Page.CaptureScreenshot(ctx, &args)

	if err != nil {
		return runtime.NewBinary([]byte{}), err
	}

	return runtime.NewBinary(reply.Data), nil
}

func (p *HTMLPage) Navigate(ctx context.Context, url runtime.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.network.Navigate(ctx, url); err != nil {
		return err
	}

	return p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) NavigateBack(ctx context.Context, skip runtime.Int) (runtime.Boolean, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret, err := p.network.NavigateBack(ctx, skip)

	if err != nil {
		return runtime.False, err
	}

	return ret, p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) NavigateForward(ctx context.Context, skip runtime.Int) (runtime.Boolean, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ret, err := p.network.NavigateForward(ctx, skip)

	if err != nil {
		return runtime.False, err
	}

	return ret, p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) WaitForNavigation(ctx context.Context, targetURL runtime.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	pattern, err := p.urlToRegexp(targetURL)

	if err != nil {
		return err
	}

	return p.waitForNavigation(ctx, cdpnet.WaitEventOptions{URL: pattern})
}

func (p *HTMLPage) WaitForFrameNavigation(ctx context.Context, frame drivers.HTMLDocument, targetURL runtime.String) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	current := p.dom.GetMainFrame()
	doc, ok := frame.(*dom.HTMLDocument)

	if !ok {
		return errors.New("invalid frame type")
	}

	pattern, err := p.urlToRegexp(targetURL)

	if err != nil {
		return err
	}

	frameID := doc.Frame().Frame.ID
	isMain := current.Frame().Frame.ID == frameID

	opts := cdpnet.WaitEventOptions{
		URL: pattern,
	}

	// if it's the current document
	if !isMain {
		opts.FrameID = frameID
	}

	return p.waitForNavigation(ctx, opts)
}

func (p *HTMLPage) Subscribe(ctx context.Context, subscription runtime.Subscription) (runtime.Stream, error) {
	switch subscription.EventName {
	case drivers.NavigationEvent:
		p.mu.Lock()
		defer p.mu.Unlock()

		stream, err := p.navigationStream(ctx)

		if err != nil {
			return nil, err
		}

		opts, err := p.parseNavigationSubscriptionOptions(ctx, subscription.Options)
		if err != nil {
			_ = stream.Close()
			return nil, err
		}

		if opts.FrameID == "" && opts.URL == nil {
			return stream, nil
		}

		return newFilteredNavigationEventStream(stream, func(evt *cdpnet.NavigationEvent) bool {
			return matchNavigationEvent(evt, opts)
		}), nil
	case drivers.RequestEvent:
		return p.network.OnRequest(ctx)
	case drivers.ResponseEvent:
		return p.network.OnResponse(ctx)
	default:
		return nil, runtime.Errorf(runtime.ErrInvalidOperation, "unknown event name: %s", subscription.EventName)
	}
}

func (p *HTMLPage) Dispatch(ctx context.Context, event runtime.DispatchEvent) error {
	return runtime.Error(runtime.ErrNotImplemented, "HTMLPage.Dispatch")
}

func (p *HTMLPage) urlToRegexp(targetURL runtime.String) (*regexp.Regexp, error) {
	if targetURL == "" {
		return nil, nil
	}

	r, err := regexp.Compile(targetURL.String())

	if err != nil {
		return nil, runtime.Error(err, "invalid URL pattern")
	}

	return r, nil
}

func (p *HTMLPage) loadMainFrame(ctx context.Context) error {
	return p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) getCurrentDocument() *dom.HTMLDocument {
	return p.dom.GetMainFrame()
}

func (p *HTMLPage) loadFrames(ctx context.Context) (runtime.List, error) {
	expectedFrames := 1

	if doc := p.getCurrentDocument(); doc != nil {
		iframeCount, err := doc.CountBySelector(ctx, drivers.NewCSSSelector("iframe"))
		if err == nil && iframeCount >= 0 {
			expectedFrames += int(iframeCount)
		}
	}

	for attempt := 0; attempt < maxFrameRefreshAttempts; attempt++ {
		if err := p.dom.ReloadRoot(ctx); err != nil {
			return nil, err
		}

		frames, err := p.dom.GetFrameNodes(ctx)
		if err != nil {
			return nil, err
		}

		count, err := frames.Length(ctx)
		if err != nil {
			return nil, err
		}

		if int(count) >= expectedFrames || attempt == maxFrameRefreshAttempts-1 {
			return frames, nil
		}

		timer := time.NewTimer(frameRefreshInterval)

		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}

			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return p.dom.GetFrameNodes(ctx)
}

func (p *HTMLPage) navigationStream(ctx context.Context) (runtime.Stream, error) {
	stream, err := p.network.OnNavigation(ctx)
	if err != nil {
		return nil, err
	}

	return newPreparedNavigationEventStream(stream, p.prepareNavigationEvent), nil
}

func (p *HTMLPage) prepareNavigationEvent(ctx context.Context, evt *cdpnet.NavigationEvent) error {
	if evt == nil {
		return nil
	}

	client := evt.SourceClient()
	if client == nil {
		client = p.client
	}

	p.dom.RecordFrameClient(evt.FrameID, client)

	exec, err := eval.Create(ctx, p.logger, client, evt.FrameID)
	if err != nil {
		return err
	}

	if _, err := events.NewEvalWaitTask(exec, templates.DOMReady(), events.DefaultPolling).Run(ctx); err != nil {
		return err
	}

	return p.dom.ReloadRoot(ctx)
}

func (p *HTMLPage) waitForNavigation(ctx context.Context, opts cdpnet.WaitEventOptions) error {
	stream, err := p.navigationStream(ctx)
	if err != nil {
		return err
	}

	defer stream.Close()

	for evt := range stream.Read(ctx) {
		if err := evt.Err(); err != nil {
			return err
		}

		nav, ok := evt.Value().(*cdpnet.NavigationEvent)
		if !ok || !matchNavigationEvent(nav, opts) {
			continue
		}

		return nil
	}

	return ctx.Err()
}

func (p *HTMLPage) parseNavigationSubscriptionOptions(ctx context.Context, value runtime.Map) (cdpnet.WaitEventOptions, error) {
	opts := cdpnet.WaitEventOptions{}
	if value == nil {
		return opts, nil
	}

	frame, err := value.Get(ctx, runtime.NewString("frame"))
	if err == nil && frame != runtime.None {
		frameID, err := navigationFrameID(frame)
		if err != nil {
			return opts, err
		}

		opts.FrameID = frameID
	}

	target, err := value.Get(ctx, runtime.NewString("target"))
	if err == nil && target != runtime.None {
		targetURL, err := runtime.CastString(target)
		if err != nil {
			return opts, err
		}

		pattern, err := p.urlToRegexp(targetURL)
		if err != nil {
			return opts, err
		}

		opts.URL = pattern
	}

	return opts, nil
}
