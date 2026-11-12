package drivers

import "net/textproto"

type (
	Options struct {
		Headers   *HTTPHeaders `json:"headers"`
		Cookies   *HTTPCookies `json:"cookies"`
		Name      string       `json:"name"`
		Proxy     string       `json:"proxy"`
		UserAgent string       `json:"userAgent"`
		// AllowedOrigins is host configuration. Nil permits all origins and an empty list denies all.
		AllowedOrigins []string `json:"-"`
	}

	Option func(opts *Options)
)

// WithAllowedOrigins sets the HTTP(S) origins the driver may request. Calling
// it with no origins enables deny-all. Each application gets its own list.
func WithAllowedOrigins(origins ...string) Option {
	values := append([]string{}, origins...)

	return func(opts *Options) {
		opts.AllowedOrigins = append([]string{}, values...)
	}
}

func WithProxy(address string) Option {
	return func(opts *Options) {
		opts.Proxy = address
	}
}

func WithUserAgent(value string) Option {
	return func(opts *Options) {
		opts.UserAgent = value
	}
}

func WithCustomName(name string) Option {
	return func(opts *Options) {
		opts.Name = name
	}
}

func WithHeader(name string, value []string) Option {
	return func(opts *Options) {
		if opts.Headers == nil {
			opts.Headers = NewHTTPHeaders()
		}

		opts.Headers.Data[name] = value
	}
}

func WithHeaders(headers textproto.MIMEHeader) Option {
	return func(opts *Options) {
		if opts.Headers == nil {
			opts.Headers = NewHTTPHeaders()
		}

		for key := range headers {
			value := headers.Get(key)

			opts.Headers.Data.Set(key, value)
		}
	}
}

func WithCookie(cookie HTTPCookie) Option {
	return func(opts *Options) {
		if opts.Cookies == nil {
			opts.Cookies = NewHTTPCookies()
		}

		opts.Cookies.Data[cookie.Name] = cookie
	}
}

func WithCookies(cookies []HTTPCookie) Option {
	return func(opts *Options) {
		if opts.Cookies == nil {
			opts.Cookies = NewHTTPCookies()
		}

		for _, c := range cookies {
			opts.Cookies.Data[c.Name] = c
		}
	}
}
