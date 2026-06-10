package jwt

import "github.com/MontFerret/contrib/modules/security/jwt/core"

const defaultMaxTokenSize = 64 * 1024

type options struct {
	maxTokenSize int
}

// Option configures SECURITY::JWT module registration.
type Option func(*options)

func newOptions(opts []Option) options {
	out := options{
		maxTokenSize: defaultMaxTokenSize,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}

	return out
}

// WithMaxTokenSize sets the maximum accepted compact JWT size in bytes.
func WithMaxTokenSize(size int) Option {
	return func(opts *options) {
		if size > 0 {
			opts.maxTokenSize = size
		}
	}
}

func (o options) coreConfig() core.Config {
	return core.Config{
		MaxTokenSize: o.maxTokenSize,
	}
}
