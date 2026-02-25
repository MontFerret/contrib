package common

import (
	"errors"
	"io"

	"github.com/MontFerret/contrib/modules/html/drivers"
	"github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/rs/zerolog"
)

var (
	ErrReadOnly    = runtime.Error(runtime.ErrInvalidOperation, "read only")
	ErrInvalidPath = runtime.Error(runtime.ErrInvalidOperation, "invalid path")
)

func CloseAll(logger zerolog.Logger, closers []io.Closer, msg string) {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			logger.Error().Err(err).Msg(msg)
		}
	}
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, runtime.ErrNotFound) || errors.Is(err, drivers.ErrNotFound)
}
