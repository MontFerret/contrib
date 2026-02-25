package drivers

type (
	globalOptions struct {
		defaultDriver string
	}

	GlobalOption func(drv Driver, opts *globalOptions)

	Options struct {
		Name      string       `json:"name"`
		Proxy     string       `json:"proxy"`
		UserAgent string       `json:"userAgent"`
		Headers   *HTTPHeaders `json:"headers"`
		Cookies   *HTTPCookies `json:"cookies"`
	}

	Option func(opts *Options)
)

func AsDefault() GlobalOption {
	return func(drv Driver, opts *globalOptions) {
		opts.defaultDriver = drv.Name()
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

func WithHeaders(headers *HTTPHeaders) Option {
	return func(opts *Options) {
		if opts.Headers == nil {
			opts.Headers = NewHTTPHeaders()
		}

		for key, _ := range headers.Data {
			if _, exists := opts.Headers.Data[key]; !exists {
				opts.Headers.Data[key] = headers.Data[key]
			}
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
