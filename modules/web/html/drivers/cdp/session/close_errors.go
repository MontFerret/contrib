package session

import (
	"context"
	"errors"
	"io"

	"github.com/mafredri/cdp/rpcc"
)

const errRPCCStreamAlreadyClosed = "rpcc: the stream is already closed"

func isIgnorableManagerStreamCloseError(err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, context.Canceled):
		return true
	case errors.Is(err, context.DeadlineExceeded):
		return true
	case errors.Is(err, io.EOF):
		return true
	case errors.Is(err, rpcc.ErrConnClosing):
		return true
	}

	return err.Error() == errRPCCStreamAlreadyClosed
}
