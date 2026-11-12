package memory

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gobwas/glob"
	"github.com/sethgrid/pester"
	"golang.org/x/net/html/charset"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
	"github.com/MontFerret/contrib/modules/web/html/drivers/internal/origin"
	"github.com/MontFerret/contrib/modules/web/html/internal/useragent"
	"github.com/MontFerret/ferret/v2/pkg/logging"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
)

const DriverName = "memory"

type Driver struct {
	client    *pester.Client
	options   *Options
	origins   *origin.Allowlist
	configErr error
}

func New(opts ...Option) *Driver {
	drv := &Driver{options: NewOptions(opts)}

	drv.origins, drv.configErr = origin.NewAllowlist(drv.options.AllowedOrigins)
	if drv.configErr != nil {
		return drv
	}

	drv.client = newHTTPClient(drv.options)
	if drv.origins != nil {
		drv.client.Transport = newOriginTransport(drv.origins, drv.client.Transport)
	}

	return drv
}

func (drv *Driver) Name() string {
	return drv.options.Name
}

func (drv *Driver) Open(ctx context.Context, params drivers.Params) (drivers.HTMLPage, error) {
	if drv.configErr != nil {
		return nil, drv.configErr
	}

	if drv.origins == nil {
		return drv.open(ctx, params)
	}

	// A denied hop cancels all retry attempts for this Open, including requests
	// cloned by the HTTP client. Keep the context alive until the body is read.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	page, err := drv.open(context.WithValue(ctx, originCancelKey{}, cancel), params)
	if cause := context.Cause(ctx); errors.Is(cause, errOriginDenied) {
		return nil, cause
	}

	return page, err
}

func (drv *Driver) open(ctx context.Context, params drivers.Params) (drivers.HTMLPage, error) {
	if params.InitScript != nil {
		return nil, runtime.Error(runtime.ErrNotSupported, "initScript is only supported by the CDP driver")
	}

	req, err := http.NewRequest(http.MethodGet, params.URL, nil)
	if err != nil {
		return nil, err
	}

	params = drivers.SetDefaultParams(drv.options.Options, params)
	req = drv.makeRequest(ctx, req, params)

	resp, err := drv.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, runtime.Errorf(err, "failed to retrieve a document %s", params.URL)
	}

	var queryFilters []drivers.StatusCodeFilter

	if params.Ignore != nil {
		queryFilters = params.Ignore.StatusCodes
	}

	if !drv.responseCodeAllowed(resp, queryFilters) {
		return nil, errors.New(resp.Status)
	}

	body := io.Reader(resp.Body)
	if params.Charset != "" {
		body, err = drv.convertToUTF8(body, params.Charset)
		if err != nil {
			return nil, runtime.Errorf(err, "failed convert to UTF-8 a document %s", params.URL)
		}
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, runtime.Errorf(err, "failed to parse a document %s", params.URL)
	}

	cookies, err := toDriverCookies(resp.Cookies())
	if err != nil {
		return nil, err
	}

	r := drivers.HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    drivers.NewHTTPHeadersWith(resp.Header),
	}

	return NewHTMLPage(doc, params.URL, r, cookies)
}

func (drv *Driver) Parse(_ context.Context, params drivers.ParseParams) (drivers.HTMLPage, error) {
	if drv.configErr != nil {
		return nil, drv.configErr
	}

	buf := bytes.NewBuffer(params.Content)

	doc, err := goquery.NewDocumentFromReader(buf)

	if err != nil {
		return nil, runtime.Errorf(err, "failed to parse a document")
	}

	return NewHTMLPage(doc, "#blank", drivers.HTTPResponse{}, nil)
}

func (drv *Driver) Close() error {
	drv.client = nil

	return nil
}

func (drv *Driver) responseCodeAllowed(resp *http.Response, additional []drivers.StatusCodeFilter) bool {
	var allowed bool
	reqURL := resp.Request.URL.String()

	// OK is by default
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true
	}

	// Try to use those that are passed within a query
	for _, filter := range additional {
		allowed = filter.Code == resp.StatusCode

		// check url
		if allowed && filter.URL != "" {
			allowed = glob.MustCompile(filter.URL).Match(reqURL)
		}

		if allowed {
			break
		}
	}

	// if still not allowed, try the default ones
	if !allowed {
		for _, filter := range drv.options.HTTPCodesFilter {
			allowed = filter.Code == resp.StatusCode

			if allowed && filter.URL != nil {
				allowed = filter.URL.Match(reqURL)
			}

			if allowed {
				break
			}
		}
	}

	return allowed
}

func (drv *Driver) convertToUTF8(reader io.Reader, srcCharset string) (data io.Reader, err error) {
	data, err = charset.NewReader(reader, srcCharset)
	if err != nil {
		return nil, err
	}

	return
}

func (drv *Driver) makeRequest(ctx context.Context, req *http.Request, params drivers.Params) *http.Request {
	logger := logging.From(ctx)

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,ru;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	if params.Headers != nil {
		headersData := params.Headers.Data

		for k := range headersData {
			v := headersData.Get(k)

			req.Header.Set(k, v)

			logger.
				Debug().
				Timestamp().
				Str("header", k).
				Msg("set header")
		}
	}

	if params.Cookies != nil {
		cookiesData := params.Cookies.Data

		for k, v := range cookiesData {
			req.AddCookie(fromDriverCookie(v))

			logger.
				Debug().
				Timestamp().
				Str("cookie", k).
				Msg("set cookie")
		}
	}

	ua := useragent.Resolve(params.UserAgent)
	logger.
		Debug().
		Timestamp().
		Str("user-agent", ua).
		Msg("using User-Agent")

	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	return req.WithContext(ctx)
}
