package memory

import (
	stdhttp "net/http"
	"slices"
	"time"

	"github.com/gobwas/glob"
	"github.com/sethgrid/pester"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

var (
	DefaultConcurrency = 1
	DefaultMaxRetries  = 5
	DefaultTimeout     = time.Second * 30
)

type (
	Option func(opts *Options)

	compiledStatusCodeFilter struct {
		URL  *glob.Pattern
		Code int
	}

	Options struct {
		*drivers.Options
		Backoff         pester.BackoffStrategy
		HTTPTransport   *stdhttp.Transport
		HTTPCodesFilter []compiledStatusCodeFilter
		MaxRetries      int
		Concurrency     int
		Timeout         time.Duration
	}
)

func NewOptions(setters []Option) *Options {
	opts := new(Options)
	opts.Options = new(drivers.Options)
	opts.Name = DriverName
	opts.Backoff = pester.ExponentialBackoff
	opts.Concurrency = DefaultConcurrency
	opts.MaxRetries = DefaultMaxRetries
	opts.Timeout = DefaultTimeout
	opts.HTTPCodesFilter = make([]compiledStatusCodeFilter, 0, 5)

	for _, setter := range setters {
		setter(opts)
	}

	return opts
}

// WithOptions applies shared driver options in order at this position in the
// constructor's option list. The supplied slice may be reused or modified.
func WithOptions(options ...drivers.Option) Option {
	setters := slices.Clone(options)

	return func(opts *Options) {
		for _, set := range setters {
			set(opts.Options)
		}
	}
}

func WithDefaultBackoff() Option {
	return func(opts *Options) {
		opts.Backoff = pester.DefaultBackoff
	}
}

func WithLinearBackoff() Option {
	return func(opts *Options) {
		opts.Backoff = pester.LinearBackoff
	}
}

func WithExponentialBackoff() Option {
	return func(opts *Options) {
		opts.Backoff = pester.ExponentialBackoff
	}
}

func WithMaxRetries(value int) Option {
	return func(opts *Options) {
		opts.MaxRetries = value
	}
}

func WithConcurrency(value int) Option {
	return func(opts *Options) {
		opts.Concurrency = value
	}
}

func WithProxy(address string) Option {
	return WithOptions(drivers.WithProxy(address))
}

func WithUserAgent(value string) Option {
	return WithOptions(drivers.WithUserAgent(value))
}

func WithCustomName(name string) Option {
	return WithOptions(drivers.WithCustomName(name))
}

func WithHeader(name string, value []string) Option {
	return WithOptions(drivers.WithHeader(name, value))
}

func WithHeaders(headers *drivers.HTTPHeaders) Option {
	return func(opts *Options) {
		if headers == nil {
			return
		}

		drivers.WithHeaders(headers.Data)(opts.Options)
	}
}

func WithCookie(cookie drivers.HTTPCookie) Option {
	return WithOptions(drivers.WithCookie(cookie))
}

func WithCookies(cookies []drivers.HTTPCookie) Option {
	return WithOptions(drivers.WithCookies(cookies))
}

// WithAllowedOrigins restricts outbound requests, including redirects and retries,
// to the listed HTTP(S) origins. An empty list denies all requests. Open and
// Parse report invalid configuration. A valid policy always permits offline Parse.
func WithAllowedOrigins(origins ...string) Option {
	return WithOptions(drivers.WithAllowedOrigins(origins...))
}

func WithAllowedHTTPCode(httpCode int) Option {
	return func(opts *Options) {
		opts.HTTPCodesFilter = append(opts.HTTPCodesFilter, compiledStatusCodeFilter{
			Code: httpCode,
		})
	}
}

func WithAllowedHTTPCodes(httpCodes []int) Option {
	return func(opts *Options) {
		for _, code := range httpCodes {
			opts.HTTPCodesFilter = append(opts.HTTPCodesFilter, compiledStatusCodeFilter{
				Code: code,
			})
		}
	}
}

func WithCustomTransport(transport *stdhttp.Transport) Option {
	return func(opts *Options) {
		opts.HTTPTransport = transport
	}
}

func WithTimeout(duration time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = duration
	}
}
