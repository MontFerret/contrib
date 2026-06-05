package sqlite

import "github.com/MontFerret/contrib/modules/db/sqlite/core"

type options struct {
	openPolicy core.OpenPolicy
}

// Option configures DB::SQLITE module registration.
type Option func(*options)

func newOptions(opts []Option) options {
	out := options{
		openPolicy: core.DefaultOpenPolicy(),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}

	return out
}

// WithMemoryOnly disables file-backed SQLite databases for registered
// DB::SQLITE::OPEN calls.
func WithMemoryOnly() Option {
	return func(opts *options) {
		opts.openPolicy = core.MemoryOnlyOpenPolicy()
	}
}
