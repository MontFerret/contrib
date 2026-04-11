package cdp

import (
	"net"
	"net/textproto"
	"net/url"
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

func WithAddress(address string) Option {
	return func(opts *Options) {
		if address != "" {
			opts.Address = normalizeAddress(address)
		}
	}
}

func WithProxy(address string) Option {
	return func(opts *Options) {
		drivers.WithProxy(address)(opts.Options)
	}
}

func WithUserAgent(value string) Option {
	return func(opts *Options) {
		drivers.WithUserAgent(value)(opts.Options)
	}
}

func WithKeepCookies() Option {
	return func(opts *Options) {
		opts.KeepCookies = true
	}
}

func WithCustomName(name string) Option {
	return func(opts *Options) {
		drivers.WithCustomName(name)(opts.Options)
	}
}

func WithHeader(name string, header []string) Option {
	return func(opts *Options) {
		drivers.WithHeader(name, header)(opts.Options)
	}
}

func WithHeaders(headers textproto.MIMEHeader) Option {
	return func(opts *Options) {
		drivers.WithHeaders(headers)(opts.Options)
	}
}

func WithCookie(cookie drivers.HTTPCookie) Option {
	return func(opts *Options) {
		drivers.WithCookie(cookie)(opts.Options)
	}
}

func WithCookies(cookies []drivers.HTTPCookie) Option {
	return func(opts *Options) {
		drivers.WithCookies(cookies)(opts.Options)
	}
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
