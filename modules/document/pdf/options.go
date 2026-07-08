package pdf

import "github.com/MontFerret/contrib/modules/document/pdf/core"

type options struct {
	openOptions core.OpenOptions
}

// Option configures DOCUMENT::PDF module registration.
type Option func(*options)

func newOptions(opts []Option) options {
	out := options{
		openOptions: core.DefaultOpenOptions(),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}

	return out
}

// WithMaxBufferSize configures the maximum PDF size that may be buffered in
// memory when Ferret's filesystem cannot provide a random-access reader.
func WithMaxBufferSize(maxBytes int64) Option {
	return func(opts *options) {
		opts.openOptions.MaxBufferSize = maxBytes
	}
}
