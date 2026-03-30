package drivers

import "net/textproto"

type (
	Options struct {
		Headers   *HTTPHeaders `json:"headers"`
		Cookies   *HTTPCookies `json:"cookies"`
		Name      string       `json:"name"`
		Proxy     string       `json:"proxy"`
		UserAgent string       `json:"userAgent"`
	}

	Option func(opts *Options)
)

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
