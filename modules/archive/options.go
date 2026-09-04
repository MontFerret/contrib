package archive

import (
	"github.com/ziflex/go-options"

	"github.com/MontFerret/contrib/modules/archive/core"
)

type Option = options.Option[core.Config]

// WithMaxEntrySize limits materialized READ values and each extracted entry.
// A zero value restores the 64 MiB default.
func WithMaxEntrySize(maxEntrySize int64) Option {
	return options.New[core.Config, int64](func(config *core.Config, value int64) {
		if value == 0 {
			value = core.DefaultConfig().MaxEntrySize
		}

		config.MaxEntrySize = value
	}).
		Named("MaxEntrySize").
		Value(maxEntrySize).
		Validators(options.NonNegative[int64]()).
		Build()
}

// WithMaxZIPBufferSize limits ZIP fallback buffering when a source provides
// neither random nor seekable access. A zero value restores the 64 MiB default.
func WithMaxZIPBufferSize(maxBytes int64) Option {
	return options.New[core.Config, int64](func(config *core.Config, value int64) {
		if value == 0 {
			value = core.DefaultConfig().MaxZIPBufferSize
		}

		config.MaxZIPBufferSize = value
	}).
		Named("MaxZIPBufferSize").
		Value(maxBytes).
		Validators(options.NonNegative[int64]()).
		Build()
}
