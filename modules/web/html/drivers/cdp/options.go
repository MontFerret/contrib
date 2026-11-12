package cdp

import (
	"net"
	"net/textproto"
	"net/url"
	"slices"
	"strings"

	"github.com/MontFerret/contrib/modules/web/html/drivers"
)

type (
	Options struct {
		*drivers.Options
		Connection  *ConnectionOptions
		Address     string
		KeepCookies bool
	}

	ConnectionOptions struct {
		BufferSize  int
		Compression bool
	}

	Option func(opts *Options)
)

const (
	DefaultAddress    = "http://localhost:9222"
	DefaultBufferSize = 1048562
)

func NewOptions(setters []Option) *Options {
	opts := new(Options)
	opts.Options = new(drivers.Options)
	opts.Name = DriverName
	opts.Address = DefaultAddress
	opts.Connection = &ConnectionOptions{
		BufferSize:  DefaultBufferSize,
		Compression: true,
	}

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

func WithAddress(address string) Option {
	return func(opts *Options) {
		if address != "" {
			opts.Address = normalizeAddress(address)
		}
	}
}

// WithAllowedOrigins restricts intercepted HTTP(S) requests across the entire
// Chromium endpoint, including unrelated browser contexts. Use a dedicated
// endpoint with one shared policy. An empty list denies every network origin.
// Configuration is copied and validated by New; Open and Parse report errors.
// The policy connection remains active until Driver.Close.
func WithAllowedOrigins(origins ...string) Option {
	return WithOptions(drivers.WithAllowedOrigins(origins...))
}

func WithProxy(address string) Option {
	return WithOptions(drivers.WithProxy(address))
}

func WithUserAgent(value string) Option {
	return WithOptions(drivers.WithUserAgent(value))
}

func WithKeepCookies() Option {
	return func(opts *Options) {
		opts.KeepCookies = true
	}
}

func WithCustomName(name string) Option {
	return WithOptions(drivers.WithCustomName(name))
}

func WithHeader(name string, header []string) Option {
	return WithOptions(drivers.WithHeader(name, header))
}

func WithHeaders(headers textproto.MIMEHeader) Option {
	return WithOptions(drivers.WithHeaders(headers))
}

func WithCookie(cookie drivers.HTTPCookie) Option {
	return WithOptions(drivers.WithCookie(cookie))
}

func WithCookies(cookies []drivers.HTTPCookie) Option {
	return WithOptions(drivers.WithCookies(cookies))
}

func WithBufferSize(size int) Option {
	return func(opts *Options) {
		opts.Connection.BufferSize = size
	}
}

func WithCompression() Option {
	return func(opts *Options) {
		opts.Connection.Compression = true
	}
}

func WithNoCompression() Option {
	return func(opts *Options) {
		opts.Connection.Compression = false
	}
}

func normalizeAddress(address string) string {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host == "" {
		return address
	}

	host := parsed.Hostname()
	if host == "" {
		return address
	}

	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return address
	}

	port := parsed.Port()
	if port == "" {
		parsed.Host = "localhost"
		return parsed.String()
	}

	parsed.Host = "localhost:" + port

	// url.URL.String may canonicalize an empty path to "", keep explicit slash form.
	if parsed.Path == "" && strings.HasSuffix(address, "/") {
		parsed.Path = "/"
	}

	return parsed.String()
}
