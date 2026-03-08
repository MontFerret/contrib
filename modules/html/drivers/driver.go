package drivers

import (
	"context"
	"io"
)

type Driver interface {
	io.Closer
	Name() string
	Open(ctx context.Context, params Params) (HTMLPage, error)
	Parse(ctx context.Context, params ParseParams) (HTMLPage, error)
}
