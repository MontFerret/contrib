package archive

import (
	"fmt"

	"github.com/ziflex/go-options"

	"github.com/MontFerret/contrib/modules/archive/core"
)

type Option = options.Option[core.Config]

// WithMaxEntrySize limits materialized READ values and each extracted entry.
// A zero value restores the 64 MiB default.
func WithMaxEntrySize(maxEntrySize int64) Option {
	return func(config *core.Config, report options.Report) {
		if maxEntrySize < 0 {
			report(options.ValidationError{
				Field:  "MaxEntrySize",
				Value:  fmt.Sprintf("%d", maxEntrySize),
				Reason: "must be non-negative",
			})

			return
		}

		config.MaxEntrySize = maxEntrySize
	}
}

// WithMaxZIPBufferSize limits ZIP fallback buffering when a source provides
// neither random nor seekable access. A zero value restores the 64 MiB default.
func WithMaxZIPBufferSize(maxBytes int64) Option {
	return func(config *core.Config, report options.Report) {
		if maxBytes < 0 {
			report(options.ValidationError{
				Field:  "MaxZIPBufferSize",
				Value:  fmt.Sprintf("%d", maxBytes),
				Reason: "must be non-negative",
			})

			return
		}

		config.MaxZIPBufferSize = maxBytes
	}
}
